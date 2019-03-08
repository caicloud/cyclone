package jsonpath

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PaesslerAG/gval"
)

//$
func getRootEvaluable(c context.Context, r, v interface{}) (interface{}, error) {
	return v, nil
}

//@
func getCurrentEvaluable(c context.Context, r, v interface{}) (interface{}, error) {
	return c.Value(currentElement{}), nil
}

type currentElement struct{}

func currentContext(c context.Context, v interface{}) context.Context {
	return context.WithValue(c, currentElement{}, v)
}

//.x, [x]
func getSelectEvaluable(key gval.Evaluable) single {
	return func(c context.Context, r, v interface{}) (interface{}, error) {

		e, _, err := selectValue(c, key, r, v)
		if err != nil {
			return nil, err
		}

		return e, nil
	}
}

// *  / [*]
func starEvaluable(c context.Context, r, v interface{}, m match) {
	visitAll(v, func(key string, val interface{}) { m(key, val) })
}

// [x, ...]
func getMultiSelectEvaluable(keys []gval.Evaluable) multi {
	if len(keys) == 0 {
		return starEvaluable
	}
	return func(c context.Context, r, v interface{}, m match) {
		for _, k := range keys {
			e, wildcard, err := selectValue(c, k, r, v)
			if err != nil {
				continue
			}
			m(wildcard, e)
		}
	}
}

func selectValue(c context.Context, key gval.Evaluable, r, v interface{}) (value interface{}, jkey string, err error) {
	c = currentContext(c, v)
	switch o := v.(type) {
	case []interface{}:
		i, err := key.EvalInt(c, r)
		if err != nil {
			return nil, "", fmt.Errorf("could not select value, invalid key: %s", err)
		}
		if i < 0 || i >= len(o) {
			return nil, "", fmt.Errorf("index %d out of bounds", i)
		}
		return o[i], strconv.Itoa(i), nil
	case map[string]interface{}:
		k, err := key.EvalString(c, r)
		if err != nil {
			return nil, "", fmt.Errorf("could not select value, invalid key: %s", err)
		}

		if r, ok := o[k]; ok {
			return r, k, nil
		}
		return nil, "", fmt.Errorf("unknown key %s", k)

	default:
		return nil, "", fmt.Errorf("unsupported value type %T for select, expected map[string]interface{} or []interface{}", o)
	}
}

//..
func mapperEvaluable(c context.Context, r, v interface{}, m match) {
	m([]interface{}{}, v)
	visitAll(v, func(wildcard string, v interface{}) {
		mapperEvaluable(c, r, v, func(key interface{}, v interface{}) {
			m(append([]interface{}{wildcard}, key.([]interface{})...), v)
		})
	})
}

func visitAll(v interface{}, visit func(key string, v interface{})) {
	switch v := v.(type) {
	case []interface{}:
		for i, e := range v {
			k := strconv.Itoa(i)
			visit(k, e)
		}
	case map[string]interface{}:
		for k, e := range v {
			visit(k, e)
		}
	}

}

//[? ]
func filterEvaluable(filter gval.Evaluable) multi {
	return func(c context.Context, r, v interface{}, m match) {
		visitAll(v, func(wildcard string, v interface{}) {
			condition, err := filter.EvalBool(currentContext(c, v), r)
			if err != nil {
				return
			}
			if condition {
				m(wildcard, v)
			}
		})
	}
}

//[::]
func getRangeEvaluable(min, max, step gval.Evaluable) multi {
	return func(c context.Context, r, v interface{}, m match) {
		cs, ok := v.([]interface{})
		if !ok {
			return
		}

		c = currentContext(c, v)

		min, err := min.EvalInt(c, r)
		if err != nil {
			return
		}
		max, err := max.EvalInt(c, r)
		if err != nil {
			return
		}
		step, err := step.EvalInt(c, r)
		if err != nil {
			return
		}

		if min > max {
			return
		}

		n := len(cs)
		min = negmax(min, n)
		max = negmax(max, n)

		if step == 0 {
			step = 1
		}

		if step > 0 {
			for i := min; i < max; i += step {
				m(strconv.Itoa(i), cs[i])
			}
		} else {
			for i := max - 1; i >= min; i += step {
				m(strconv.Itoa(i), cs[i])
			}
		}

	}
}

func negmax(n, max int) int {
	if n < 0 {
		n = max + n
		if n < 0 {
			n = 0
		}
	} else if n > max {
		return max
	}
	return n
}

// ()
func newScript(script gval.Evaluable) single {
	return func(c context.Context, r, v interface{}) (interface{}, error) {
		return script(currentContext(c, v), r)
	}
}
