package gval

import (
	"context"
	"fmt"
	"text/scanner"
	"unicode"
)

// Language is an expression language
type Language struct {
	prefixes  map[interface{}]prefix
	operators map[string]operator
}

// NewLanguage returns the union of given Languages as new Language.
func NewLanguage(bases ...Language) Language {
	l := newLanguage()
	for _, base := range bases {
		for i, e := range base.prefixes {
			l.prefixes[i] = e
		}
		for i, e := range base.operators {
			l.operators[i] = e.merge(l.operators[i])
			l.operators[i].initiate(i)
		}
	}
	return l
}

func newLanguage() Language {
	return Language{
		prefixes:  map[interface{}]prefix{},
		operators: map[string]operator{},
	}
}

// NewEvaluable returns an Evaluable for given expression in the specified language
func (l Language) NewEvaluable(expression string) (Evaluable, error) {
	p := newParser(expression, l)

	eval, err := p.ParseExpression(context.Background())

	if err == nil && p.isCamouflaged() && p.lastScan != scanner.EOF {
		err = p.camouflage
	}

	if err != nil {
		pos := p.scanner.Pos()
		return nil, fmt.Errorf("parsing error: %s - %d:%d %s", p.scanner.Position, pos.Line, pos.Column, err)
	}
	return eval, nil
}

// Evaluate given parameter with given expression
func (l Language) Evaluate(expression string, parameter interface{}) (interface{}, error) {
	eval, err := l.NewEvaluable(expression)
	if err != nil {
		return nil, err
	}
	return eval(context.Background(), parameter)
}

// Function returns a Language with given function.
// Function has no conversion for input types.
//
// If the function returns an error it must be the last return parameter.
//
// If the function has (without the error) more then one return parameter,
// it returns them as []interface{}.
func Function(name string, function interface{}) Language {
	l := newLanguage()
	l.prefixes[name] = func(c context.Context, p *Parser) (eval Evaluable, err error) {
		args := []Evaluable{}
		scan := p.Scan()
		switch scan {
		case '(':
			args, err = p.parseArguments(c)
			if err != nil {
				return nil, err
			}
		default:
			p.Camouflage("function call", '(')
		}
		return p.callFunc(toFunc(function), args...), nil
	}
	return l
}

// Constant returns a Language with given constant
func Constant(name string, value interface{}) Language {
	l := newLanguage()
	l.prefixes[name] = func(c context.Context, p *Parser) (eval Evaluable, err error) {
		return p.Const(value), nil
	}
	return l
}

// PrefixExtension extends a Language
func PrefixExtension(r rune, ext func(context.Context, *Parser) (Evaluable, error)) Language {
	l := newLanguage()
	l.prefixes[r] = ext
	return l
}

// PrefixMetaPrefix chooses a Prefix to be executed
func PrefixMetaPrefix(r rune, ext func(context.Context, *Parser) (call string, alternative func() (Evaluable, error), err error)) Language {
	l := newLanguage()
	l.prefixes[r] = func(c context.Context, p *Parser) (Evaluable, error) {
		call, alternative, err := ext(c, p)
		if err != nil {
			return nil, err
		}
		key := interface{}(call)
		if len(call) == 1 && !unicode.IsLetter(([]rune(call))[0]) {
			key = ([]rune(call))[0] //TODO getter and setter
		}
		if prefix, ok := p.prefixes[key]; ok {
			return prefix(c, p)
		}
		return alternative()
	}
	return l
}

//PrefixOperator returns a Language with given prefix
func PrefixOperator(name string, e Evaluable) Language {
	l := newLanguage()
	key := interface{}(name)
	if len(name) == 1 && !unicode.IsLetter(([]rune(name))[0]) {
		key = ([]rune(name))[0] //TODO getter and setter
	}
	l.prefixes[key] = func(c context.Context, p *Parser) (Evaluable, error) {
		eval, err := p.ParseNextExpression(c)
		if err != nil {
			return nil, err
		}
		prefix := func(c context.Context, v interface{}) (interface{}, error) {
			a, err := eval(c, v)
			if err != nil {
				return nil, err
			}
			return e(c, a)
		}
		if eval.IsConst() {
			v, err := prefix(context.Background(), nil)
			if err != nil {
				return nil, err
			}
			prefix = p.Const(v)
		}
		return prefix, nil
	}
	return l
}

// PostfixOperator extends a Language
func PostfixOperator(name string, ext func(context.Context, *Parser, Evaluable) (Evaluable, error)) Language {
	l := newLanguage()
	l.operators[name] = postfix{
		f: func(c context.Context, p *Parser, eval Evaluable, pre operatorPrecedence) (Evaluable, error) {
			return ext(c, p, eval)
		},
	}
	return l
}

// InfixOperator for two arbitrary values.
func InfixOperator(name string, f func(a, b interface{}) (interface{}, error)) Language {
	return newLanguageOperator(name, &infix{arbitrary: f})
}

// InfixShortCircuit operator is called after the left operand is evaluated.
func InfixShortCircuit(name string, f func(a interface{}) (interface{}, bool)) Language {
	return newLanguageOperator(name, &infix{shortCircuit: f})
}

// InfixTextOperator for two text values.
func InfixTextOperator(name string, f func(a, b string) (interface{}, error)) Language {
	return newLanguageOperator(name, &infix{text: f})
}

// InfixNumberOperator for two number values.
func InfixNumberOperator(name string, f func(a, b float64) (interface{}, error)) Language {
	return newLanguageOperator(name, &infix{number: f})
}

// InfixBoolOperator for two bool values.
func InfixBoolOperator(name string, f func(a, b bool) (interface{}, error)) Language {
	return newLanguageOperator(name, &infix{boolean: f})
}

// Precedence of operator. The Operator with higher operatorPrecedence is evaluated first.
func Precedence(name string, operatorPrecendence uint8) Language {
	return newLanguageOperator(name, operatorPrecedence(operatorPrecendence))
}

// InfixEvalOperator operates on the raw operands.
// Therefore it cannot be combined with operators for other operand types.
func InfixEvalOperator(name string, f func(a, b Evaluable) (Evaluable, error)) Language {
	return newLanguageOperator(name, directInfix{infixBuilder: f})
}

func newLanguageOperator(name string, op operator) Language {
	op.initiate(name)
	l := newLanguage()
	l.operators[name] = op
	return l
}
