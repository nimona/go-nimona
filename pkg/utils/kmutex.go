package utils

import (
	"sync"
)

// Kmutex is a key based mutex, allowing us to lock/unlock specific keys
type Kmutex struct {
	*sync.Map
}

// NewKmutex creates a new Kmutex
func NewKmutex() Kmutex {
	return Kmutex{
		&sync.Map{},
	}
}

// Unlock a key
func (km *Kmutex) Unlock(key interface{}) {
	imutex, exist := km.Load(key)
	if !exist {
		panic("kmutex: unlock of unlocked mutex")
	}
	mutex := imutex.(*sync.Mutex)
	km.Delete(key)
	mutex.Unlock()
}

// Lock a key
func (km *Kmutex) Lock(key interface{}) {
	nmutex := &sync.Mutex{}
	imutex, _ := km.LoadOrStore(key, nmutex)
	mutex := imutex.(*sync.Mutex)
	mutex.Lock()
	if mutex != nmutex {
		mutex.Unlock()
		km.Lock(key)
		return
	}
	return
}
