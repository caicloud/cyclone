package kubernetes

import (
	"fmt"
	"sync"
)

type ClusterClientsetGetter interface {
	Get(cluster string) Interface
	Put(cluster string, kc Interface)
	Del(cluster string)
	Range(f func(cluster string, kc Interface) error) error
}

type ClusterClientsetMap struct {
	sm sync.Map
}

func NewClusterClientsetGetter(m map[string]Interface) ClusterClientsetGetter {
	var sm sync.Map
	for k, v := range m {
		if v != nil {
			sm.Store(k, v)
		}
	}
	return &ClusterClientsetMap{sm}
}
func (m ClusterClientsetMap) Get(cluster string) Interface {
	loaded, exist := m.sm.Load(cluster)
	if !exist || loaded == nil {
		return nil
	}
	c, exist := loaded.(Interface)
	if !exist {
		return nil
	}
	return c
}
func (m ClusterClientsetMap) Put(cluster string, kc Interface) {
	if kc != nil {
		m.sm.Store(cluster, kc)
	}
}
func (m ClusterClientsetMap) Del(cluster string) {
	m.sm.Delete(cluster)
}
func (m ClusterClientsetMap) Range(f func(cluster string, kc Interface) error) (e error) {
	m.sm.Range(func(key, value interface{}) bool {
		cluster, ok := key.(string)
		if !ok {
			e = fmt.Errorf("bad key, no cluster name inside")
			return false
		}
		kc, ok := value.(Interface)
		if !ok || kc == nil {
			e = fmt.Errorf("bad value %s, no cluster client inside", cluster)
			return false
		}
		e = f(cluster, kc)
		if e != nil {
			return false
		}
		return true
	})
	return e
}
