package jsonpath

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"text/scanner"

	"github.com/PaesslerAG/gval"
)

type jsonObject []func(c context.Context, v interface{}, visit func(key string, value interface{})) error

type keyValuePair struct {
	key   gval.Evaluable
	value gval.Evaluable
}

type keyValueMatcher struct {
	key     gval.Evaluable
	matcher func(c context.Context, r, v interface{}, visit func(keys []interface{}, match interface{}))
}

func parseJSONObject(c context.Context, p *gval.Parser) (gval.Evaluable, error) {
	evals := jsonObject{}
	for {
		switch p.Scan() {
		default:
			hasWildcard := false

			p.Camouflage("object", ',', '}')
			key, err := p.ParseExpression(context.WithValue(c, hasPlaceholdersContextKey{}, &hasWildcard))
			if err != nil {
				return nil, err
			}
			if p.Scan() != ':' {
				if err != nil {
					return nil, p.Expected("object", ':')
				}
			}
			switch hasWildcard {
			case true:
				jp := &parser{Parser: p}
				switch p.Scan() {
				case '$':
					jp.single = getRootEvaluable
				case '@':
					jp.single = getCurrentEvaluable
				default:
					return nil, p.Expected("JSONPath key and value")
				}
				err = jp.parsePath(c)

				if err != nil {
					return nil, err
				}
				evals = append(evals, keyValueMatcher{key, jp.getMultis().visitMatchs}.visit)
			case false:
				value, err := p.ParseExpression(c)
				if err != nil {
					return nil, err
				}
				evals = append(evals, keyValuePair{key, value}.visit)
			}
		case ',':
		case '}':
			return evals.evaluable, nil
		}
	}
}

func (kv keyValuePair) visit(c context.Context, v interface{}, visit func(key string, value interface{})) error {
	value, err := kv.value(c, v)
	if err != nil {
		return err
	}
	key, err := kv.key.EvalString(c, v)
	if err != nil {
		return err
	}
	visit(key, value)
	return nil
}

func (kv keyValueMatcher) visit(c context.Context, v interface{}, visit func(key string, value interface{})) (err error) {
	kv.matcher(c, v, v, func(keys []interface{}, match interface{}) {
		key, er := kv.key.EvalString(context.WithValue(c, placeholdersContextKey{}, keys), v)
		if er != nil {
			err = er
		}
		visit(key, match)
	})
	return
}

func (j jsonObject) evaluable(c context.Context, v interface{}) (interface{}, error) {
	vs := map[string]interface{}{}
	for _, e := range j {
		err := e(c, v, func(key string, value interface{}) { vs[key] = value })
		if err != nil {
			return nil, err
		}
	}
	return vs, nil
}

func parsePlaceholder(c context.Context, p *gval.Parser) (gval.Evaluable, error) {
	hasWildcard := c.Value(hasPlaceholdersContextKey{})
	if hasWildcard == nil {
		return nil, fmt.Errorf("JSONPath placeholder must only be used in an JSON object key")
	}
	*(hasWildcard.(*bool)) = true
	switch p.Scan() {
	case scanner.Int:
		id, err := strconv.Atoi(p.TokenText())
		if err != nil {
			return nil, err
		}
		return placeholder(id).evaluable, nil
	default:
		p.Camouflage("JSONPath placeholder")
		return allPlaceholders.evaluable, nil
	}
}

type hasPlaceholdersContextKey struct{}

type placeholdersContextKey struct{}

type placeholder int

const allPlaceholders = placeholder(-1)

func (key placeholder) evaluable(c context.Context, v interface{}) (interface{}, error) {
	wildcards, ok := c.Value(placeholdersContextKey{}).([]interface{})
	if !ok || len(wildcards) <= int(key) {
		return nil, fmt.Errorf("JSONPath placeholder #%d is not available", key)
	}
	if key == allPlaceholders {
		sb := bytes.Buffer{}
		sb.WriteString("$")
		quoteWildcardValues(&sb, wildcards)
		return sb.String(), nil
	}
	return wildcards[int(key)], nil
}

func quoteWildcardValues(sb *bytes.Buffer, wildcards []interface{}) {
	for _, w := range wildcards {
		if wildcards, ok := w.([]interface{}); ok {
			quoteWildcardValues(sb, wildcards)
			continue
		}
		sb.WriteString(fmt.Sprintf("[%v]",
			strconv.Quote(fmt.Sprint(w)),
		))
	}
}
