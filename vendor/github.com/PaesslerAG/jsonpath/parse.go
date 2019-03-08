package jsonpath

import (
	"context"
	"fmt"
	"math"
	"text/scanner"

	"github.com/PaesslerAG/gval"
)

func (s single) parse(c context.Context, p *gval.Parser) (r gval.Evaluable, err error) {
	jp := &parser{Parser: p, single: s}
	err = jp.parsePath(c)

	if err != nil {
		return nil, err
	}
	return jp.evaluable(), nil
}

func (jp *parser) parsePath(c context.Context) error {
	switch jp.Scan() {
	case '.':
		return jp.parseSelect(c)
	case '[':
		keys, seperator, err := jp.parseBracket(c)

		if err != nil {
			return err
		}

		switch seperator {
		case ':':
			if len(keys) > 3 {
				return fmt.Errorf("range query has at least the parameter [min:max:step]")
			}
			keys = append(keys, []gval.Evaluable{
				jp.Const(0), jp.Const(float64(math.MaxInt32)), jp.Const(1)}[len(keys):]...)
			jp.newMultiStage(getRangeEvaluable(keys[0], keys[1], keys[2]))
		case '?':
			if len(keys) != 1 {
				return fmt.Errorf("filter needs exactly one key")
			}
			jp.newMultiStage(filterEvaluable(keys[0]))
		default:
			if len(keys) == 1 {
				jp.newSingleStage(getSelectEvaluable(keys[0]))
			} else {
				jp.newMultiStage(getMultiSelectEvaluable(keys))
			}
		}
		return jp.parsePath(c)
	case '(':
		return jp.parseScript(c)
	default:
		jp.Camouflage("jsonpath", '.', '[', '(')
		return nil
	}
}

func (jp *parser) parseSelect(c context.Context) error {
	scan := jp.Scan()
	switch scan {
	case scanner.Ident:
		jp.newSingleStage(getSelectEvaluable(jp.Const(jp.TokenText())))
		return jp.parsePath(c)
	case '.':
		jp.newMultiStage(mapperEvaluable)
		return jp.parseMapper(c)
	case '*':
		jp.newMultiStage(starEvaluable)
		return jp.parsePath(c)
	default:
		return jp.Expected("JSON select", scanner.Ident, '.', '*')
	}
}

func (jp *parser) parseBracket(c context.Context) (keys []gval.Evaluable, seperator rune, err error) {
	for {
		scan := jp.Scan()
		skipScan := false
		switch scan {
		case '?':
			skipScan = true
		case ':':
			i := float64(0)
			if len(keys) == 1 {
				i = math.MaxInt32
			}
			keys = append(keys, jp.Const(i))
			skipScan = true
		case '*':
			if jp.Scan() != ']' {
				return nil, 0, jp.Expected("JSON bracket star", ']')
			}
			return []gval.Evaluable{}, 0, nil
		case ']':
			if seperator == ':' {
				skipScan = true
				break
			}
			fallthrough
		default:
			jp.Camouflage("jsonpath brackets")
			key, err := jp.ParseExpression(c)
			if err != nil {
				return nil, 0, err
			}
			keys = append(keys, key)
		}
		if !skipScan {
			scan = jp.Scan()
		}
		if seperator == 0 {
			seperator = scan
		}
		switch scan {
		case ':', ',':
		case ']':
			return
		case '?':
			if len(keys) != 0 {
				return nil, 0, jp.Expected("JSON filter", ']')
			}
		default:
			return nil, 0, jp.Expected("JSON bracket separator", ':', ',')
		}
		if seperator != scan {
			return nil, 0, fmt.Errorf("mixed %v and %v in JSON bracket", seperator, scan)
		}
	}
}

func (jp *parser) parseMapper(c context.Context) error {
	scan := jp.Scan()
	switch scan {
	case scanner.Ident:
		jp.newSingleStage(getSelectEvaluable(jp.Const(jp.TokenText())))
	case '[':
		keys, seperator, err := jp.parseBracket(c)

		if err != nil {
			return err
		}
		switch seperator {
		case ':':
			return fmt.Errorf("mapper can not be combined with range query")
		case '?':
			if len(keys) != 1 {
				return fmt.Errorf("filter needs exactly one key")
			}
			jp.newMultiStage(filterEvaluable(keys[0]))
		default:
			jp.newMultiStage(getMultiSelectEvaluable(keys))
		}
	case '*':
		jp.newMultiStage(starEvaluable)
	case '(':
		return jp.parseScript(c)
	default:
		return jp.Expected("JSON mapper", '[', scanner.Ident, '*')
	}
	return jp.parsePath(c)
}

func (jp *parser) parseScript(c context.Context) error {
	script, err := jp.ParseExpression(c)
	if err != nil {
		return err
	}
	if jp.Scan() != ')' {
		return jp.Expected("jsnopath script", ')')
	}
	jp.newSingleStage(newScript(script))
	return jp.parsePath(c)
}
