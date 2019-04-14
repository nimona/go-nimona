package sync

import (
	"sync"
)

// Kmutex is a key based mutex, allowing us to lock/unlock specific keys
type Kmutex struct {
	m *sync.Map
}

// NewKmutex creates a new Kmutex
func NewKmutex() Kmutex {
	return Kmutex{
		m: &sync.Map{},
	}
}

// Unlock a key
func (km *Kmutex) Unlock(key interface{}) {
	imutex, exist := km.m.Load(key)
	if !exist {
		panic("kmutex: unlock of unlocked mutex")
	}
	mutex := imutex.(*sync.Mutex)
	km.m.Delete(key)
	mutex.Unlock()
}

// Lock a key
func (km *Kmutex) Lock(key interface{}) {
	nmutex := &sync.Mutex{}
	imutex, _ := km.m.LoadOrStore(key, nmutex)
	mutex := imutex.(*sync.Mutex)
	mutex.Lock()
	if mutex != nmutex {
		mutex.Unlock()
		km.Lock(key)
		return
	}
	return
}
