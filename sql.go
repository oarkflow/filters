package filters

import (
	"errors"
	"fmt"
	"strings"
)

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
	switch strings.ToUpper(s) {
	case "SELECT", "FROM", "WHERE", "AND", "OR", "IS", "NULL", "NOT", "BETWEEN", "LIKE":
		return true
	default:
		return false
	}
}

func isOperator(s string) bool {
	switch strings.ToUpper(s) {
	case "=", "!=", ">", "<", ">=", "<=", "LIKE", "NOT LIKE", "BETWEEN", "IN":
		return true
	default:
		return false
	}
}

func tokenize(input string) ([]token, error) {
	var tokens []token

	isWhitespace := func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n'
	}

	isDigit := func(r rune) bool {
		return r >= '0' && r <= '9'
	}

	input = strings.TrimSpace(input)
	for i := 0; i < len(input); {
		r := rune(input[i])
		switch {
		case isWhitespace(r):
			i++
		case r == '(':
			tokens = append(tokens, token{typ: tokenLParen, value: string(r)})
			i++
		case r == ')':
			tokens = append(tokens, token{typ: tokenRParen, value: string(r)})
			i++
		case r == ',':
			tokens = append(tokens, token{typ: tokenComma, value: string(r)})
			i++
		case r == '\'':
			start := i + 1
			i++
			for i < len(input) && rune(input[i]) != '\'' {
				i++
			}
			if i < len(input) && rune(input[i]) == '\'' {
				tokens = append(tokens, token{typ: tokenValue, value: input[start:i]})
				i++
			} else {
				return nil, errors.New("unclosed string literal")
			}
		case isDigit(r):
			start := i
			for i < len(input) && isDigit(rune(input[i])) {
				i++
			}
			tokens = append(tokens, token{typ: tokenValue, value: input[start:i]})
		case isOperator(string(r)):
			start := i
			for i < len(input) && isOperator(string(input[i])) {
				i++
			}
			tokens = append(tokens, token{typ: tokenOperator, value: input[start:i]})
		default:
			start := i
			for i < len(input) && !isWhitespace(rune(input[i])) && !isOperator(string(input[i])) && input[i] != '(' && input[i] != ')' && input[i] != ',' {
				i++
			}
			value := input[start:i]
			if isKeyword(value) {
				tokens = append(tokens, token{typ: tokenKeyword, value: strings.ToUpper(value)})
			} else if strings.ToUpper(value) == "TRUE" || strings.ToUpper(value) == "FALSE" {
				tokens = append(tokens, token{typ: tokenValue, value: strings.ToUpper(value)})
			} else {
				tokens = append(tokens, token{typ: tokenIdentifier, value: value})
			}
		}
	}

	// Combine tokens to correctly form compound operators like "IS NULL" and "IS NOT NULL"
	for i := 0; i < len(tokens)-1; i++ {
		if tokens[i].typ == tokenKeyword && strings.ToUpper(tokens[i].value) == "IS" {
			if i+1 < len(tokens) && tokens[i+1].typ == tokenKeyword {
				if strings.ToUpper(tokens[i+1].value) == "NULL" {
					tokens[i] = token{typ: tokenOperator, value: "IS NULL"}
					tokens = append(tokens[:i+1], tokens[i+2:]...)
				} else if strings.ToUpper(tokens[i+1].value) == "NOT" && i+2 < len(tokens) && strings.ToUpper(tokens[i+2].value) == "NULL" {
					tokens[i] = token{typ: tokenOperator, value: "IS NOT NULL"}
					tokens = append(tokens[:i+1], tokens[i+3:]...)
				}
			}
		}
	}

	return tokens, nil
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

func mapSQLOperatorToGoOperator(sqlOperator string) Operator {
	switch strings.ToUpper(sqlOperator) {
	case "=":
		return Equal
	case "!=":
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

func parseFilter(tokens []token) (*Filter, int, error) {
	p := &parser{tokens: tokens}

	tok, ok := p.nextToken()
	if !ok || tok.typ != tokenIdentifier {
		return nil, 0, errors.New("expected field name")
	}
	field := tok.value

	tok, ok = p.nextToken()
	if !ok || tok.typ != tokenOperator {
		fmt.Println(tok)
		return nil, 0, errors.New("expected operator")
	}
	operator := mapSQLOperatorToGoOperator(tok.value)

	var value any
	if operator != IsNull && operator != NotNull {
		tok, ok = p.nextToken()
		if !ok || (tok.typ != tokenValue && tok.typ != tokenIdentifier) {
			return nil, 0, errors.New("expected value")
		}
		value = tok.value
	}

	filter := &Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}

	return filter, p.pos, nil
}

func parseFilterGroup(tokens []token) (*FilterGroup, int, error) {
	p := &parser{tokens: tokens}
	var filters []*Filter
	var operator Boolean

	for {
		tok, ok := p.peekToken()
		if !ok {
			break
		}

		if tok.typ == tokenBoolean {
			p.nextToken() // consume the boolean token
			operator = Boolean(strings.ToUpper(tok.value))
			continue
		}

		if tok.typ == tokenLParen {
			p.nextToken() // consume '('
			group, consumed, err := parseFilterGroup(tokens[p.pos:])
			if err != nil {
				return nil, 0, err
			}
			p.pos += consumed

			filters = append(filters, &Filter{
				Field:    "",
				Operator: "",
				Value:    group,
			})
			continue
		}

		if tok.typ == tokenRParen {
			p.nextToken() // consume ')'
			break
		}

		filter, consumed, err := parseFilter(tokens[p.pos:])
		if err != nil {
			return nil, 0, err
		}
		filters = append(filters, filter)

		p.pos += consumed
	}

	return &FilterGroup{
		Operator: operator,
		Filters:  filters,
	}, p.pos, nil
}

func FromSQL(sql string) (*FilterGroup, error) {
	tokens, err := tokenize(sql)
	if err != nil {
		return nil, err
	}
	fmt.Println(tokens)
	var whereTokens []token
	whereClause := false

	for _, tok := range tokens {
		if tok.typ == tokenKeyword && strings.ToUpper(tok.value) == "WHERE" {
			whereClause = true
			continue
		}
		if whereClause || tok.typ != tokenKeyword {
			whereTokens = append(whereTokens, tok)
		}
	}

	if len(whereTokens) == 0 {
		return nil, errors.New("no WHERE clause found")
	}

	filterGroup, _, err := parseFilterGroup(whereTokens)
	if err != nil {
		return nil, err
	}
	return filterGroup, nil
}
