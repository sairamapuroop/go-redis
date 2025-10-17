package db

import (
	"sync"
)

type DB struct {
	mu    sync.RWMutex
	store map[string]string
}

func New() *DB {
	return &DB{
		store: make(map[string]string),
	}
}

func (d *DB) Get(key string) (string, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	val, ok := d.store[key]

	return val, ok
}

func (d *DB) Set(key string, val string) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	d.store[key] = val
}

func (d *DB) Delete(key string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	_, ok := d.store[key]

	if ok {
		delete(d.store, key)
	}

	return ok

}
