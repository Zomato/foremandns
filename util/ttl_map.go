package util

import (
	"sync"
	"time"
)

type item struct {
	value      string
	lastAccess int64
}

// TTLMap map with expire functionality
type TTLMap struct {
	m map[string]*item
	l sync.Mutex
}

// New TTLMap interface
func New(maxTTL int) (m *TTLMap) {
	m = &TTLMap{m: map[string]*item{}}
	go func() {
		for now := range time.Tick(time.Second) {
			m.l.Lock()
			for k, v := range m.m {
				if now.Unix()-v.lastAccess > int64(maxTTL) {
					delete(m.m, k)
				}
			}
			m.l.Unlock()
		}
	}()
	return
}

// Len length of the TTLMap
func (m *TTLMap) Len() int {
	return len(m.m)
}

// Put set the value
func (m *TTLMap) Put(k, v string) {
	m.l.Lock()
	it, ok := m.m[k]
	if !ok {
		it = &item{value: v}
		m.m[k] = it
	}
	it.lastAccess = time.Now().Unix()
	m.l.Unlock()
}

// Get get the value
func (m *TTLMap) Get(k string) (v string) {
	m.l.Lock()
	if it, ok := m.m[k]; ok {
		v = it.value
		it.lastAccess = time.Now().Unix()
	}
	m.l.Unlock()
	return

}
