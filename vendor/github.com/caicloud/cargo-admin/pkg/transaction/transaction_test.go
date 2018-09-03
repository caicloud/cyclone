package transaction

import (
	"fmt"
	"testing"

	"github.com/caicloud/cargo-admin/pkg/errors"
)

func TestNormal(t *testing.T) {
	state := make(map[string]int)
	transaction := New()
	action := &Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			state["a"] = 0
			return nil, nil
		},
		Rollbacker: func(args ...interface{}) error {
			delete(state, "a")
			return nil
		},
	}
	transaction.Add(action)

	action = &Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			state["a"]++
			return nil, nil
		},
		Rollbacker: func(args ...interface{}) error {
			state["a"]--
			return nil
		},
	}
	transaction.Add(action)
	transaction.Run()

	if v, ok := state["a"]; !ok || v != 1 {
		fmt.Errorf("Expect %d, but got %d", 1, v)
	}
}

func TestError(t *testing.T) {
	state := make(map[string]int)
	transaction := New()
	action := &Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			state["a"] = 1
			return nil, nil
		},
		Rollbacker: func(args ...interface{}) error {
			delete(state, "a")
			return nil
		},
	}
	transaction.Add(action)

	action = &Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			return nil, errors.ErrorUnknownInternal.Error("intended")
		},
		Rollbacker: func(args ...interface{}) error {
			return nil
		},
	}
	transaction.Add(action)
	transaction.Run()

	if _, ok := state["a"]; ok {
		fmt.Errorf("Expected key %s not exists, but found exist", "a")
	}
}

func TestRollbackerArgs(t *testing.T) {
	state := make(map[string]int)
	transaction := New()
	action := &Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			state["a"] = 1
			state["b"] = 1
			state["c"] = 1
			return []interface{}{"a", "b", "c"}, nil
		},
		Rollbacker: func(args ...interface{}) error {
			for _, arg := range args {
				delete(state, arg.(string))
			}
			return nil
		},
	}
	transaction.Add(action)

	action = &Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			return nil, errors.ErrorUnknownInternal.Error("intended")
		},
		Rollbacker: func(args ...interface{}) error {
			return nil
		},
	}
	transaction.Add(action)
	transaction.Run()

	if len(state) > 0 {
		fmt.Errorf("Expected state to be empty, but got %v", state)
	}
}

func TestRunnerArgs(t *testing.T) {
	state := make(map[string]int)
	transaction := New()
	action := &Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			state["d"] = 4
			return []interface{}{[]string{"a", "b", "c"}, []int{1, 2, 3}}, nil
		},
		Rollbacker: func(args ...interface{}) error {
			delete(state, "d")
			return nil
		},
	}
	transaction.Add(action)

	action = &Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			keys := args[0].([]string)
			values := args[1].([]int)
			for i, k := range keys {
				state[k] = values[i]
			}
			return []interface{}{keys}, nil
		},
		Rollbacker: func(args ...interface{}) error {
			for _, k := range args[0].([]string) {
				delete(state, k)
			}
			return nil
		},
	}
	transaction.Add(action)

	transaction.Run()
	keys := []string{"a", "b", "c", "d"}
	values := []int{1, 2, 3, 4}
	ok := true
	for i, k := range keys {
		if state[k] != values[i] {
			ok = false
		}
	}
	if !ok {
		t.Errorf("%v not expected", state)
	}
}
