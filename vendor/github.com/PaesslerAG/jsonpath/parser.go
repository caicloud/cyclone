package jsonpath

import (
	"context"

	"github.com/PaesslerAG/gval"
)

type parser struct {
	*gval.Parser
	single
	multis multis
}

//multi evaluate wildcard
type multi func(c context.Context, r, v interface{}, m match)

type multis []multi

type match func(key, v interface{})

//single evaluate exactly one result
type single func(c context.Context, r, v interface{}) (interface{}, error)

func (p *parser) newSingleStage(next single) {
	if p.single == nil {
		p.single = next
		return
	}
	last := p.single
	p.single = func(c context.Context, r, v interface{}) (interface{}, error) {
		v, err := last(c, r, v)
		if err != nil {
			return nil, err
		}
		return next(c, r, v)
	}
}

func (p *parser) newMultiStage(next multi) {
	if p.single != nil {
		s := p.single
		p.single = nil
		p.multis = append(p.multis, func(c context.Context, r, v interface{}, m match) {
			v, err := s(c, r, v)
			if err != nil {
				return
			}
			next(c, r, v, m)
		})
		return
	}
	p.multis = append(p.multis, next)
}

func (p *parser) evaluable() gval.Evaluable {
	if p.multis == nil {
		return p.single.evaluable
	}
	return p.getMultis().evaluable
}

func (p *parser) getMultis() multis {
	last := len(p.multis) - 1
	if p.single != nil {
		p.multis[last] = p.multis[last].append(p.single)
		p.single = nil
	}
	return p.multis
}

func (ms multis) visitMatchs(c context.Context, r, v interface{}, visit func(keys []interface{}, match interface{})) {
	if len(ms) == 0 {
		visit(nil, v)
		return
	}
	ms[0](c, r, v, func(key, v interface{}) {
		ms[1:].visitMatchs(c, r, v, func(keys []interface{}, match interface{}) {
			visit(append([]interface{}{key}, keys...), match)
		})
	})
}

func (ms multis) evaluable(c context.Context, v interface{}) (interface{}, error) {
	matchs := []interface{}{}
	ms.visitMatchs(c, v, v, func(keys []interface{}, match interface{}) {
		matchs = append(matchs, match)
	})
	return matchs, nil
}

func (s single) evaluable(c context.Context, v interface{}) (interface{}, error) {
	return s(c, v, v)
}

func (multi multi) append(s single) multi {
	return func(c context.Context, r, v interface{}, m match) {
		multi(c, r, v, func(key, v interface{}) {
			v, err := s(c, r, v)
			if err != nil {
				return
			}
			m(key, v)
		})
	}
}
