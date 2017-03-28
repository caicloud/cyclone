package register

import "sync"

// Register is a struct binds name and interface such as Constructor
type Register struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewRegister returns a new register
func NewRegister() *Register {
	return &Register{
		data: make(map[string]interface{}),
	}
}

// Register binds name and interface
// It will panic if name already exists
func (r *Register) Register(name string, v interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.data[name]
	if ok {
		panic("Repeated registration key: " + name)
	}
	r.data[name] = v

}

// Get returns an interface registered with the given name
func (r *Register) Get(name string) interface{} {
	// need lock ?
	return r.data[name]
}
