package transaction

import "github.com/caicloud/nirvana/log"

// Transaction wraps execution of a series of Actions, when one action failed,
// all previous actions will be rolled back by the provided Rollbacker function.
type Transaction struct {
	// Actions to be executed in order, if one Action failed, the remaining will not be
	// executed and previous ones will be rolled back.
	actions []*Action

	// Results of each action executed, they will be accessed in the rollback function
	results [][]interface{}
}

// Each Action should contains only one action that changes state and this
// action should be executed as later as possible. For example, when you
// have several retrieval actions and one write action, try to put the write
// action in the last.
// Runner: func() ([]interface{}, error) {
//	err := get()
//	if err != nil {
//		return err
//	}
//	err = getMore()
//	if err != nil {
//		return err
//	}
//	return write()
//}
// When you add actions in a for loop, mind the closure problem. As the sample below, if
// contextObj is not declared, all action will refer to the last obj
//for _, obj := range objs {
//	contextObj := obj
//	action := &Action{
//		Runner: func() ([]interface{}, error) {
//			fmt.Println(contextObj)
//		}
//	}
//}
// Runner has arguments "args" that are the runner results of previous action in the transaction
// Rollbacker has arguments "args" that are the runner results of current action
type Action struct {
	Runner     func(args ...interface{}) ([]interface{}, error)
	Rollbacker func(args ...interface{}) error
}

func New() *Transaction {
	return &Transaction{
		actions: make([]*Action, 0),
		results: make([][]interface{}, 0),
	}
}

func (t *Transaction) Add(action *Action) {
	t.actions = append(t.actions, action)
}

func (t *Transaction) Run() error {
	failed := -1
	var err error
	for i, a := range t.actions {
		var args []interface{}
		if i > 0 {
			args = t.results[i-1]
		}
		r, e := a.Runner(args...)

		t.results = append(t.results, r)
		if e != nil {
			failed = i
			err = e
			break
		}
	}
	if failed >= 0 {
		for i := failed - 1; i >= 0; i-- {
			e := t.actions[i].Rollbacker(t.results[i]...)
			if e != nil {
				log.Errorf("Rollback action error: %v", e)
				break
			}
		}
		return err
	}

	return nil
}
