package kmutex

import (
	"sync"
)

// Kmutex mutex by unique ID
type Kmutex struct {
	c *sync.Cond
	l sync.Locker
	s map[interface{}]struct{}
}

// New creates a new Kmutex
func New() *Kmutex {
	l := sync.Mutex{}
	return &Kmutex{c: sync.NewCond(&l), l: &l, s: make(map[interface{}]struct{})}
}

func (km *Kmutex) locked(key interface{}) (ok bool) { _, ok = km.s[key]; return }

// Unlock Kmutex by unique ID
func (km *Kmutex) Unlock(key interface{}) {
	km.l.Lock()
	defer km.l.Unlock()
	delete(km.s, key)
	km.c.Broadcast()
}

// Lock Kmutex by unique ID
func (km *Kmutex) Lock(key interface{}) {
	km.l.Lock()
	defer km.l.Unlock()
	for km.locked(key) {
		km.c.Wait()
	}
	km.s[key] = struct{}{}
	return
}
