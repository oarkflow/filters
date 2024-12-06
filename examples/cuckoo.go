package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func main() {
	docs := []Document{
		{ID: 1, Data: map[string]string{"title": "Go Programming", "content": "Learn Go with examples", "category": "Programming"}},
		{ID: 2, Data: map[string]string{"title": "Full Text Search", "content": "Efficient search with BM25", "category": "Search"}},
		{ID: 3, Data: map[string]string{"title": "Fuzzy Search", "content": "Handle typos and errors", "category": "Search"}},
	}
	engine := &FullTextSearchEngine{
		InvertedIndex: NewInvertedIndex(),
		BKTree:        NewBKTree(),
		BM25Params:    BM25Parameters{K1: 1.5, B: 0.75},
	}
	for _, doc := range docs {
		engine.InvertedIndex.AddDocument(doc)
		for _, field := range doc.Data {
			for _, token := range Tokenize(field) {
				engine.BKTree.AddTerm(token)
			}
		}
	}
	filters := []Filter{
		{Field: "category", Operator: "=", Value: "Search"},
	}
	results := engine.Search("search", nil, 1, filters)
	fmt.Printf("Filtered Results: %+v\n", results)
}

type Document struct {
	ID   int
	Data map[string]string
}

type InvertedIndex struct {
	Index map[string]map[string][]int
	Docs  map[int]Document
}

type BKTreeNode struct {
	Term     string
	Children map[int]*BKTreeNode
}

type BKTree struct {
	Root *BKTreeNode
}

type BM25Parameters struct {
	K1 float64
	B  float64
}

type FullTextSearchEngine struct {
	InvertedIndex *InvertedIndex
	BKTree        *BKTree
	BM25Params    BM25Parameters
}

type Filter struct {
	Field    string
	Operator string
	Value    any
}

func applyFilter(doc Document, filter Filter) bool {
	fieldValue, exists := doc.Data[filter.Field]
	if !exists {
		return false
	}
	fieldValueStr := fieldValue
	switch filter.Operator {
	case "=":
		return fieldValueStr == filter.Value
	case "!=":
		return fieldValueStr != filter.Value
	case ">":

		if val, ok := filter.Value.(float64); ok {
			fieldValFloat, err := strconv.ParseFloat(fieldValueStr, 64)
			if err == nil {
				return fieldValFloat > val
			}
		}
	case "<":
		if val, ok := filter.Value.(float64); ok {
			fieldValFloat, err := strconv.ParseFloat(fieldValueStr, 64)
			if err == nil {
				return fieldValFloat < val
			}
		}
	default:
		return false
	}
	return false
}

func Tokenize(data string) []string {
	return strings.Fields(strings.ToLower(data))
}

func (idx *InvertedIndex) AddDocument(doc Document) {
	idx.Docs[doc.ID] = doc
	for field, content := range doc.Data {
		tokens := Tokenize(content)
		for _, token := range tokens {
			if _, ok := idx.Index[token]; !ok {
				idx.Index[token] = make(map[string][]int)
			}
			idx.Index[token][field] = append(idx.Index[token][field], doc.ID)
		}
	}
}

func (idx *InvertedIndex) SearchExact(queryTokens []string, fields []string) []int {
	results := map[int]bool{}
	for _, token := range queryTokens {
		if tokenFields, ok := idx.Index[token]; ok {
			if fields == nil {
				for _, docIDs := range tokenFields {
					for _, docID := range docIDs {
						results[docID] = true
					}
				}
			} else {
				for _, field := range fields {
					if docIDs, ok := tokenFields[field]; ok {
						for _, docID := range docIDs {
							results[docID] = true
						}
					}
				}
			}
		}
	}
	var resultIDs []int
	for docID := range results {
		resultIDs = append(resultIDs, docID)
	}
	return resultIDs
}

func ComputeBM25(query []string, idx *InvertedIndex, params BM25Parameters) map[int]float64 {
	scores := make(map[int]float64)
	N := len(idx.Docs)
	var totalDocLength int
	docLengths := make(map[int]int)
	for id, doc := range idx.Docs {
		length := 0
		for _, field := range doc.Data {
			length += len(Tokenize(field))
		}
		docLengths[id] = length
		totalDocLength += length
	}
	avgDocLength := float64(totalDocLength) / float64(N)
	for _, term := range query {
		if tokenFields, ok := idx.Index[term]; ok {
			docs := map[int]bool{}
			for _, docIDs := range tokenFields {
				for _, docID := range docIDs {
					docs[docID] = true
				}
			}
			idf := math.Log(1 + (float64(N)-float64(len(docs))+0.5)/(float64(len(docs))+0.5))
			for docID := range docs {
				tf := float64(len(idx.Index[term]))
				length := float64(docLengths[docID])
				score := idf * ((tf * (params.K1 + 1)) / (tf + params.K1*(1-params.B+params.B*(length/avgDocLength))))
				scores[docID] += score
			}
		}
	}
	return scores
}

func (tree *BKTree) AddTerm(term string) {
	if tree.Root == nil {
		tree.Root = &BKTreeNode{Term: term, Children: make(map[int]*BKTreeNode)}
		return
	}
	current := tree.Root
	for {
		distance := EditDistance(term, current.Term)
		if child, exists := current.Children[distance]; exists {
			current = child
		} else {
			current.Children[distance] = &BKTreeNode{Term: term, Children: make(map[int]*BKTreeNode)}
			return
		}
	}
}

func (tree *BKTree) SearchFuzzy(query string, maxDistance int) []string {
	var results []string
	var search func(node *BKTreeNode, query string, maxDistance int)
	search = func(node *BKTreeNode, query string, maxDistance int) {
		distance := EditDistance(node.Term, query)
		if distance <= maxDistance {
			results = append(results, node.Term)
		}
		for d, child := range node.Children {
			if d >= distance-maxDistance && d <= distance+maxDistance {
				search(child, query, maxDistance)
			}
		}
	}
	if tree.Root != nil {
		search(tree.Root, query, maxDistance)
	}
	return results
}

func (fts *FullTextSearchEngine) Search(query string, fields []string, fuzzyThreshold int, filters []Filter) []Document {
	queryTokens := Tokenize(query)
	exactMatches := fts.InvertedIndex.SearchExact(queryTokens, fields)
	var fuzzyTokens []string
	for _, token := range queryTokens {
		fuzzyTokens = append(fuzzyTokens, fts.BKTree.SearchFuzzy(token, fuzzyThreshold)...)
	}
	matches := map[int]bool{}
	for _, id := range exactMatches {
		matches[id] = true
	}
	for _, token := range fuzzyTokens {
		for field, docIDs := range fts.InvertedIndex.Index[token] {
			if fields == nil || contains(fields, field) {
				for _, docID := range docIDs {
					matches[docID] = true
				}
			}
		}
	}
	var filteredMatches []int
	for docID := range matches {
		doc := fts.InvertedIndex.Docs[docID]
		includeDoc := true
		for _, filter := range filters {
			if !applyFilter(doc, filter) {
				includeDoc = false
				break
			}
		}
		if includeDoc {
			filteredMatches = append(filteredMatches, docID)
		}
	}
	scores := ComputeBM25(queryTokens, fts.InvertedIndex, fts.BM25Params)
	var results []struct {
		ID    int
		Score float64
	}
	for _, docID := range filteredMatches {
		results = append(results, struct {
			ID    int
			Score float64
		}{ID: docID, Score: scores[docID]})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	var filteredDocs []Document
	for _, result := range results {
		filteredDocs = append(filteredDocs, fts.InvertedIndex.Docs[result.ID])
	}
	return filteredDocs
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func EditDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
	}
	for i := 0; i <= len(a); i++ {
		dp[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		dp[0][j] = j
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = int(math.Min(float64(dp[i-1][j]+1),
				math.Min(float64(dp[i][j-1]+1), float64(dp[i-1][j-1]+cost))))
		}
	}
	return dp[len(a)][len(b)]
}

func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		Index: make(map[string]map[string][]int),
		Docs:  make(map[int]Document),
	}
}

func NewBKTree() *BKTree {
	return &BKTree{}
}
