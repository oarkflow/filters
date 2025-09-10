package filters

import (
	"errors"
	"slices"
	"strings"
)

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
		return NewFilter(field, NotEndsWith, val), "", p.pos, nil
	case strings.HasSuffix(tok.value, "%"):
		return NewFilter(field, NotStartsWith, val), "", p.pos, nil
	}
	return nil, "", 0, errors.New("unexpected LIKE pattern")
}
