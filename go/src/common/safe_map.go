package common

import (
	"sync"
)

type SafeMap struct {
	sync.RWMutex
	m map[interface{}]interface{}
}

func NewSafeMap() *SafeMap {
	m := new(SafeMap)
	m.m = make(map[interface{}]interface{})
	return m
}
func (m *SafeMap) Get(key interface{}) interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.m[key]
}
func (m *SafeMap) Set(key interface{}, value interface{}) {
	m.Lock()
	defer m.Unlock()
	m.m[key] = value
}
func (m *SafeMap) Del(key interface{}) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, key)
}
func (m *SafeMap) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.m)
}

func (m *SafeMap) RLockRange(f func(interface{}, interface{})) {
	m.RLock()
	defer m.RUnlock()
	for k, v := range m.m {
		f(k, v)
	}
}
func (m *SafeMap) LockRange(f func(interface{}, interface{})) {
	m.Lock()
	defer m.Unlock()
	for k, v := range m.m {
		f(k, v)
	}
}
