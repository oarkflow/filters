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
	switch strings.ToUpper(s) {
	case "SELECT", "FROM", "WHERE", "AND", "OR", "IS", "NULL", "NOT LIKE", "NOT", "BETWEEN", "LIKE", "IN":
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

func toOperator(sqlOperator string) Operator {
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
