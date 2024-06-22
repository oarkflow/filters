package filters

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

type Condition interface {
	Match(data any) bool
}

type Sequence struct {
	Node     Condition
	Operator Boolean
	Next     Condition
	Result   bool
}

func (s *Sequence) Match(data any) bool {
	matched := s.Node.Match(data)
	if s.Operator == AND && !matched {
		return false
	}
	if s.Next == nil {
		return matched
	}
	matchedNext := s.Next.Match(data)
	if s.Operator == AND {
		return matched && matchedNext
	} else if s.Operator == OR {
		return matched || matchedNext
	}
	return matchedNext
}

func FilterCondition[T any](data []T, expr *Sequence) (result []T) {
	for _, d := range data {
		if expr.Match(d) {
			result = append(result, d)
		}
	}
	return
}

type tokenType string

const (
	tokenIdentifier tokenType = "IDENTIFIER"
	tokenOperator   tokenType = "OPERATOR"
	tokenValue      tokenType = "VALUE"
	tokenBoolean    tokenType = "BOOLEAN"
	tokenLParen     tokenType = "LPAREN"
	tokenRParen     tokenType = "RPAREN"
	tokenKeyword    tokenType = "KEYWORD"
	tokenComma      tokenType = "COMMA"
)

type token struct {
	typ   tokenType
	value string
}

func isKeyword(s string) bool {
	keywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "AND": true, "OR": true,
		"IS": true, "NULL": true, "NOT LIKE": true, "NOT": true, "BETWEEN": true,
		"LIKE": true, "IN": true,
	}
	return keywords[strings.ToUpper(s)]
}

func isOperator(s string) bool {
	operators := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
		"LIKE": true, "NOT LIKE": true, "BETWEEN": true, "IN": true,
	}
	return operators[strings.ToUpper(s)]
}

func tokenize(input string) ([]token, error) {
	var tokens []token
	input = strings.TrimSpace(input)

	for i := 0; i < len(input); {
		switch r := input[i]; {
		case isWhitespace(r):
			i++
		case r == '(', r == ')', r == ',':
			tokens = append(tokens, token{typ: runeToTokenType(r), value: string(r)})
			i++
		case r == '\'':
			value, newIndex, err := parseStringLiteral(input, i)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token{typ: tokenValue, value: value})
			i = newIndex
		case isDigit(r):
			value, newIndex := parseNumber(input, i)
			tokens = append(tokens, token{typ: tokenValue, value: value})
			i = newIndex
		case isOperatorStart(r):
			value, newIndex := parseOperator(input, i)
			tokens = append(tokens, token{typ: tokenOperator, value: value})
			i = newIndex
		default:
			value, newIndex := parseIdentifierOrKeyword(input, i)
			if isKeyword(value) {
				tokens = append(tokens, token{typ: tokenKeyword, value: strings.ToUpper(value)})
			} else if isBoolean(value) {
				tokens = append(tokens, token{typ: tokenValue, value: strings.ToUpper(value)})
			} else {
				tokens = append(tokens, token{typ: tokenIdentifier, value: value})
			}
			i = newIndex
		}
	}

	tokens = combineCompoundOperators(tokens)
	return tokens, nil
}

func runeToTokenType(r byte) tokenType {
	switch r {
	case '(':
		return tokenLParen
	case ')':
		return tokenRParen
	case ',':
		return tokenComma
	default:
		return ""
	}
}

func isWhitespace(r byte) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

func isDigit(r byte) bool {
	return r >= '0' && r <= '9'
}

func isOperatorStart(r byte) bool {
	return isOperator(string(r))
}

func parseStringLiteral(input string, start int) (string, int, error) {
	i := start + 1
	for i < len(input) && input[i] != '\'' {
		i++
	}
	if i < len(input) && input[i] == '\'' {
		return input[start+1 : i], i + 1, nil
	}
	return "", 0, errors.New("unclosed string literal")
}

func parseNumber(input string, start int) (string, int) {
	i := start
	for i < len(input) && isDigit(input[i]) {
		i++
	}
	return input[start:i], i
}

func parseOperator(input string, start int) (string, int) {
	i := start
	for i < len(input) && isOperator(string(input[i])) {
		i++
	}
	return input[start:i], i
}

func parseIdentifierOrKeyword(input string, start int) (string, int) {
	i := start
	for i < len(input) && !isWhitespace(input[i]) && !isOperatorStart(input[i]) && input[i] != '(' && input[i] != ')' && input[i] != ',' {
		i++
	}
	return input[start:i], i
}

func isBoolean(s string) bool {
	return strings.ToUpper(s) == "TRUE" || strings.ToUpper(s) == "FALSE"
}

func combineCompoundOperators(tokens []token) []token {
	var result []token
	for i := 0; i < len(tokens); i++ {
		if tokens[i].typ == tokenKeyword && strings.ToUpper(tokens[i].value) == "IS" {
			if i+1 < len(tokens) && tokens[i+1].typ == tokenKeyword {
				if strings.ToUpper(tokens[i+1].value) == "NULL" {
					result = append(result, token{typ: tokenOperator, value: "IS NULL"})
					i++
					continue
				} else if strings.ToUpper(tokens[i+1].value) == "NOT" && i+2 < len(tokens) && strings.ToUpper(tokens[i+2].value) == "NULL" {
					result = append(result, token{typ: tokenOperator, value: "IS NOT NULL"})
					i += 2
					continue
				}
			}
		}
		result = append(result, tokens[i])
	}
	return result
}

type parser struct {
	tokens []token
	pos    int
}

func (p *parser) nextToken() (token, bool) {
	if p.pos >= len(p.tokens) {
		return token{}, false
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok, true
}

func (p *parser) peekToken() (token, bool) {
	if p.pos >= len(p.tokens) {
		return token{}, false
	}
	return p.tokens[p.pos], true
}

func toOperator(sqlOperator string) Operator {
	switch strings.ToUpper(sqlOperator) {
	case "=":
		return Equal
	case "!=", "<>":
		return NotEqual
	case ">":
		return GreaterThan
	case "<":
		return LessThan
	case ">=":
		return GreaterThanEqual
	case "<=":
		return LessThanEqual
	case "LIKE":
		return Contains
	case "NOT LIKE":
		return NotContains
	case "BETWEEN":
		return Between
	case "IN":
		return In
	case "IS NULL":
		return IsNull
	case "IS NOT NULL":
		return NotNull
	default:
		return ""
	}
}

func parseFilter(tokens []token) (*Filter, Boolean, int, error) {
	p := &parser{tokens: tokens}

	tok, ok := p.nextToken()
	if !ok || (tok.typ == tokenKeyword && slices.Contains([]Boolean{AND, OR}, Boolean(tok.value))) {
		return nil, Boolean(tok.value), p.pos, nil
	}
	if !ok || tok.typ != tokenIdentifier {
		return nil, "", 0, errors.New("expected field name")
	}
	field := tok.value
	tok, ok = p.nextToken()
	if !ok {
		return nil, "", 0, errors.New("unexpected error1")
	}
	switch tok.typ {
	case tokenOperator:
		operator := toOperator(tok.value)

		var value any
		if operator != IsNull && operator != NotNull {
			tok, ok = p.nextToken()
			if !ok || (tok.typ != tokenValue && tok.typ != tokenIdentifier) {
				return nil, "", 0, errors.New("expected value")
			}
			value = tok.value
		}

		return NewFilter(field, operator, value), "", p.pos, nil
	case tokenKeyword:
		if tok.value == "BETWEEN" {
			operator := toOperator(tok.value)
			tok, ok = p.nextToken()
			if !ok || tok.typ != tokenIdentifier {
				return nil, "", 0, errors.New("expected field name")
			}
			field1 := tok.value
			tok, ok = p.nextToken()
			if !ok || tok.value != "AND" {
				return nil, "", 0, errors.New("expected AND Operator")
			}
			tok, ok = p.nextToken()
			if !ok || tok.typ != tokenIdentifier {
				return nil, "", 0, errors.New("expected field name")
			}
			field2 := tok.value
			return NewFilter(field, operator, []any{field1, field2}), "", p.pos, nil
		} else if tok.value == "LIKE" {
			tok, ok = p.nextToken()
			if !ok || tok.typ != tokenIdentifier {
				return nil, "", 0, errors.New("expected field name")
			}
			if strings.HasPrefix(tok.value, "%") && strings.HasSuffix(tok.value, "%") {
				val := strings.ReplaceAll(tok.value, "%", "")
				return NewFilter(field, Contains, val), "", p.pos, nil
			} else if strings.HasPrefix(tok.value, "%") && !strings.HasSuffix(tok.value, "%") {
				val := strings.ReplaceAll(tok.value, "%", "")
				return NewFilter(field, StartsWith, val), "", p.pos, nil
			} else if !strings.HasPrefix(tok.value, "%") && strings.HasSuffix(tok.value, "%") {
				val := strings.ReplaceAll(tok.value, "%", "")
				return NewFilter(field, EndsWith, val), "", p.pos, nil
			}
		} else if tok.value == "IN" {
			var in []any
			tok, ok = p.nextToken()
			if !ok && tok.typ != tokenLParen {
				return nil, "", 0, errors.New("unexpected error")
			}

			for {
				tok, _ = p.nextToken()
				if tok.typ == tokenRParen {
					break
				}
				if tok.typ == tokenValue || tok.typ == tokenIdentifier {
					in = append(in, tok.value)
				}
			}
			return NewFilter(field, In, in), "", p.pos, nil
		} else if tok.value == "NOT" {
			tok, ok = p.nextToken()
			if !ok {
				return nil, "", 0, errors.New("unexpected error")
			}
			switch tok.value {
			case "IN":
				var in []any
				tok, ok = p.nextToken()
				if !ok && tok.typ != tokenLParen {
					return nil, "", 0, errors.New("unexpected error")
				}

				for {
					tok, _ = p.nextToken()
					if tok.typ == tokenRParen {
						break
					}
					if tok.typ == tokenValue || tok.typ == tokenIdentifier {
						in = append(in, tok.value)
					}
				}
				return NewFilter(field, NotIn, in), "", p.pos, nil
			case "LIKE":
				tok, ok = p.nextToken()
				if !ok || tok.typ != tokenIdentifier {
					return nil, "", 0, errors.New("expected field name")
				}
				if strings.HasPrefix(tok.value, "%") && strings.HasSuffix(tok.value, "%") {
					val := strings.ReplaceAll(tok.value, "%", "")
					filter := NewFilter(field, NotContains, val)
					return filter, "", p.pos, nil
				} else if strings.HasPrefix(tok.value, "%") && !strings.HasSuffix(tok.value, "%") {
					val := strings.ReplaceAll(tok.value, "%", "")
					filter := &Filter{
						Field:    field,
						Operator: StartsWith,
						Value:    val,
						Reverse:  true,
					}
					return filter, "", p.pos, nil
				} else if !strings.HasPrefix(tok.value, "%") && strings.HasSuffix(tok.value, "%") {
					val := strings.ReplaceAll(tok.value, "%", "")
					filter := &Filter{
						Field:    field,
						Operator: EndsWith,
						Value:    val,
						Reverse:  true,
					}
					return filter, "", p.pos, nil
				}
			}
		}
	}
	return nil, "", 0, errors.New("unexpected error")
}

func parseFilterGroup(tokens []token) (*Sequence, int, error) {
	p := &parser{tokens: tokens}
	seq := &Sequence{}
	var operator Boolean

	for {
		tok, ok := p.peekToken()
		if !ok {
			break
		}

		if tok.typ == tokenBoolean {
			p.nextToken() // consume the boolean token
			operator = Boolean(strings.ToUpper(tok.value))
			seq.Operator = operator
			continue
		}

		if tok.typ == tokenLParen {
			p.nextToken() // consume '('
			group, consumed, err := parseFilterGroup(tokens[p.pos:])
			if err != nil {
				return nil, 0, err
			}
			p.pos += consumed
			if seq.Node == nil {
				seq.Node = group
			} else {
				seq.Next = group
			}
			continue
		}

		if tok.typ == tokenRParen {
			p.nextToken() // consume ')'
			break
		}
		filter, ops, consumed, err := parseFilter(tokens[p.pos:])
		if err != nil {
			return nil, 0, err
		}
		p.pos += consumed
		if ops != "" {
			seq.Operator = ops
		} else {
			if seq.Node == nil {
				seq.Node = filter
			} else {
				seq.Next = filter
			}
		}
	}
	return seq, p.pos, nil
}

func FirstTermFilter(seq *Sequence) (*Filter, error) {
	if seq == nil {
		return nil, errors.New("sequence is nil")
	}

	// Helper function to traverse the sequence recursively
	var traverse func(node any) *Filter
	traverse = func(node any) *Filter {
		switch n := node.(type) {
		case *Filter:
			if n.Operator == Equal {
				return n
			}
		case *Sequence:
			if n.Node != nil {
				if filter := traverse(n.Node); filter != nil {
					return filter
				}
			}
			if n.Next != nil {
				if filter := traverse(n.Next); filter != nil {
					return filter
				}
			}
		}
		return nil
	}

	// Start traversing from the root of the sequence
	if filter := traverse(seq.Node); filter != nil {
		return filter, nil
	}
	if filter := traverse(seq.Next); filter != nil {
		return filter, nil
	}

	return nil, errors.New("no equal filter found")
}

func ParseSQL(sql string) (*Sequence, error) {
	sql = splitByWhere(sql)
	tokens, err := tokenize(sql)
	if err != nil {
		return nil, err
	}

	filterGroup, _, err := parseFilterGroup(tokens)
	if err != nil {
		return nil, err
	}
	return filterGroup, nil
}

var (
	re = regexp.MustCompile(`(?i)\bWHERE\b`)
)

func splitByWhere(sql string) string {
	loc := re.FindStringIndex(sql)
	if loc == nil {
		return sql
	}

	beforeWhere := strings.TrimSpace(sql[:loc[0]])
	afterWhere := strings.TrimSpace(sql[loc[1]:])
	if afterWhere != "" {
		return afterWhere
	}
	return beforeWhere
}
