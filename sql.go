package filters

import (
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/oarkflow/filters/utils"
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
	tokenVariable   tokenType = "VARIABLE"
)

type token struct {
	typ   tokenType
	value string
}

var (
	keywords = map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "AND": true, "OR": true,
		"IS": true, "NULL": true, "NOT LIKE": true, "NOT": true, "BETWEEN": true, "NOT BETWEEN": true,
		"LIKE": true, "IN": true,
	}
	operators = map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
		"LIKE": true, "NOT LIKE": true, "BETWEEN": true, "NOT BETWEEN": true, "IN": true, "NOT IN": true,
		"IS NULL": true, "IS NOT NULL": true,
	}
)

func isKeyword(s string) bool {
	return keywords[strings.ToUpper(s)]
}

func isOperator(s string) bool {
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
		case r == '{' && i+1 < len(input) && input[i+1] == '{':
			value, newIndex, err := parseVariable(input, i)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token{typ: tokenVariable, value: "{{" + value + "}}"})
			i = newIndex
		case isDigit(r):
			value, newIndex := parseIdentifierOrKeyword(input, i)
			if utils.IsValidDateTime(value) {
				tokens = append(tokens, token{typ: tokenValue, value: strings.ToUpper(value)})
			} else {
				value, newIndex = parseNumber(input, i)
				tokens = append(tokens, token{typ: tokenValue, value: value})
			}
			i = newIndex
		case isOperatorStart(r):
			value, newIndex := parseOperator(input, i)
			tokens = append(tokens, token{typ: tokenOperator, value: value})
			i = newIndex
			value, newIndex = parseIdentifierOrKeyword(input, i)
			if utils.IsValidDateTime(value) {
				tokens = append(tokens, token{typ: tokenValue, value: strings.ToUpper(value)})
				i = newIndex
			}
		default:
			value, newIndex := parseIdentifierOrKeyword(input, i)
			if isKeyword(value) {
				tokens = append(tokens, token{typ: tokenKeyword, value: strings.ToUpper(value)})
			} else if utils.IsValidDateTime(value) {
				tokens = append(tokens, token{typ: tokenValue, value: strings.ToUpper(value)})
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

func parseVariable(input string, start int) (string, int, error) {
	i := start + 2
	for i < len(input) && !(input[i] == '}' && i+1 < len(input) && input[i+1] == '}') {
		i++
	}
	if i < len(input) && input[i] == '}' && i+1 < len(input) && input[i+1] == '}' {
		return input[start+2 : i], i + 2, nil
	}
	return "", 0, errors.New("unclosed variable")
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
		if tokens[i].typ == tokenKeyword && strings.ToUpper(tokens[i].value) == "NOT" {
			if i+1 < len(tokens) && tokens[i+1].typ == tokenKeyword {
				compoundOperator := strings.ToUpper(tokens[i].value + " " + tokens[i+1].value)
				if isOperator(compoundOperator) {
					result = append(result, token{typ: tokenOperator, value: compoundOperator})
					i++
					continue
				}
			}
		}
		if tokens[i].typ == tokenKeyword && strings.ToUpper(tokens[i].value) == "IS" {
			if i+1 < len(tokens) && tokens[i+1].typ == tokenKeyword {
				compoundOperator := strings.ToUpper(tokens[i].value + " " + tokens[i+1].value)
				if isOperator(compoundOperator) {
					result = append(result, token{typ: tokenOperator, value: compoundOperator})
					i++
					continue
				}
				if strings.ToUpper(tokens[i+1].value) == "NOT" && i+2 < len(tokens) && tokens[i+2].typ == tokenKeyword {
					compoundOperator = strings.ToUpper(tokens[i].value + " " + tokens[i+1].value + " " + tokens[i+2].value)
					if isOperator(compoundOperator) {
						result = append(result, token{typ: tokenOperator, value: compoundOperator})
						i += 2
						continue
					}
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
	if tok.typ != tokenIdentifier {
		return nil, "", 0, errors.New("expected field name")
	}
	field := tok.value

	tok, ok = p.nextToken()
	if !ok {
		return nil, "", 0, errors.New("unexpected error")
	}
	if tok.typ == tokenOperator {
		operator := toOperator(tok.value)
		var value any
		if operator != IsNull && operator != NotNull {
			tok, ok = p.nextToken()
			if !ok || (tok.typ != tokenValue && tok.typ != tokenIdentifier && tok.typ != tokenVariable) {
				return nil, "", 0, errors.New("expected value")
			}
			value = tok.value
		}
		if operator == NotContains {
			val := value.(string)
			val = strings.Trim(val, "%")
			switch {
			case strings.HasPrefix(val, "%") && strings.HasSuffix(val, "%"):
				return NewFilter(field, NotContains, val), "", p.pos, nil
			case strings.HasPrefix(val, "%"):
				return NewFilter(field, NotEndsWith, val), "", p.pos, nil
			case strings.HasSuffix(val, "%"):
				return NewFilter(field, NotStartsWith, val), "", p.pos, nil
			}
		}
		return NewFilter(field, operator, value), "", p.pos, nil
	}

	if tok.typ == tokenKeyword {
		switch tok.value {
		case "BETWEEN":
			return parseBetween(p, field)
		case "LIKE":
			return parseLike(p, field)
		case "IN":
			return parseIn(p, field)
		case "NOT":
			return parseNot(p, field)
		}
	}

	return nil, "", 0, errors.New("unexpected token")
}

func parseBetween(p *parser, field string) (*Filter, Boolean, int, error) {
	operator := Between

	tok, ok := p.nextToken()
	if !ok || !slices.Contains([]tokenType{tokenIdentifier, tokenValue, tokenVariable}, tok.typ) {
		return nil, "", 0, errors.New("expected value")
	}
	field1 := tok.value

	tok, ok = p.nextToken()
	if !ok || tok.value != "AND" {
		return nil, "", 0, errors.New("expected AND operator")
	}

	tok, ok = p.nextToken()
	if !ok || !slices.Contains([]tokenType{tokenIdentifier, tokenValue, tokenVariable}, tok.typ) {
		return nil, "", 0, errors.New("expected value")
	}
	field2 := tok.value

	return NewFilter(field, operator, []any{field1, field2}), "", p.pos, nil
}

func parseLike(p *parser, field string) (*Filter, Boolean, int, error) {
	tok, ok := p.nextToken()
	if !ok || !slices.Contains([]tokenType{tokenIdentifier, tokenValue, tokenVariable}, tok.typ) {
		return nil, "", 0, errors.New("expected value")
	}
	val := strings.Trim(tok.value, "%")
	switch {
	case strings.HasPrefix(tok.value, "%") && strings.HasSuffix(tok.value, "%"):
		return NewFilter(field, Contains, val), "", p.pos, nil
	case strings.HasPrefix(tok.value, "%"):
		return NewFilter(field, EndsWith, val), "", p.pos, nil
	case strings.HasSuffix(tok.value, "%"):
		return NewFilter(field, StartsWith, val), "", p.pos, nil
	}
	return nil, "", 0, errors.New("unexpected LIKE pattern")
}

func parseIn(p *parser, field string) (*Filter, Boolean, int, error) {
	var in []any
	tok, ok := p.nextToken()
	if !ok || tok.typ != tokenLParen {
		return nil, "", 0, errors.New("expected '('")
	}
	for {
		tok, _ = p.nextToken()
		if tok.typ == tokenRParen {
			break
		}
		if slices.Contains([]tokenType{tokenIdentifier, tokenValue, tokenVariable}, tok.typ) {
			in = append(in, tok.value)
		}
	}
	return NewFilter(field, In, in), "", p.pos, nil
}

func parseNot(p *parser, field string) (*Filter, Boolean, int, error) {
	tok, ok := p.nextToken()
	if !ok {
		return nil, "", 0, errors.New("unexpected error")
	}
	switch tok.value {
	case "IN":
		return parseNotIn(p, field)
	case "LIKE":
		return parseNotLike(p, field)
	}
	return nil, "", 0, errors.New("unexpected NOT token")
}

func parseNotIn(p *parser, field string) (*Filter, Boolean, int, error) {
	var in []any
	tok, ok := p.nextToken()
	if !ok || tok.typ != tokenLParen {
		return nil, "", 0, errors.New("expected '('")
	}
	for {
		tok, _ = p.nextToken()
		if tok.typ == tokenRParen {
			break
		}
		if slices.Contains([]tokenType{tokenIdentifier, tokenValue, tokenVariable}, tok.typ) {
			in = append(in, tok.value)
		}
	}
	return NewFilter(field, NotIn, in), "", p.pos, nil
}

func parseNotLike(p *parser, field string) (*Filter, Boolean, int, error) {
	tok, ok := p.nextToken()
	if !ok || tok.typ != tokenIdentifier {
		return nil, "", 0, errors.New("expected value")
	}
	val := strings.Trim(tok.value, "%")
	switch {
	case strings.HasPrefix(tok.value, "%") && strings.HasSuffix(tok.value, "%"):
		return NewFilter(field, NotContains, val), "", p.pos, nil
	case strings.HasPrefix(tok.value, "%"):
		return &Filter{Field: field, Operator: EndsWith, Value: val, Reverse: true}, "", p.pos, nil
	case strings.HasSuffix(tok.value, "%"):
		return &Filter{Field: field, Operator: StartsWith, Value: val, Reverse: true}, "", p.pos, nil
	}
	return nil, "", 0, errors.New("unexpected LIKE pattern")
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
