package db

import (
	"log"
	"sync"
	"time"
)

type item struct {
	value string
	expiresAt time.Time
}

type DB struct {
	mu    sync.RWMutex
	store map[string]item
}

func New() *DB {
	return &DB{
		store: make(map[string]item),
	}
}

func (d *DB) Get(key string) (string, bool) {
	  d.mu.RLock()
    itm, ok := d.store[key]
    d.mu.RUnlock()

    if !ok {
        return "", false
    }

    if itm.expiresAt.IsZero() || itm.expiresAt.After(time.Now()) {
        return itm.value, true
    }

    // expired, remove it
    d.mu.Lock()
    delete(d.store, key)
    d.mu.Unlock()
    return "", false
}

func (d *DB) Set(key, val string, ttl time.Duration) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var expiresAt time.Time

	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}


	d.store[key] = item{
		value: val,
		expiresAt: expiresAt,
	}
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

func (d *DB) Flush() {
	d.mu.Lock()
	d.store = make(map[string]item)
	d.mu.Unlock()
}

func (d *DB) StartJanitor(interval time.Duration, stopCh <-chan struct{}) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                d.cleanup()
            case <-stopCh:
                return
            }
        }
    }()
}

func (d *DB) cleanup() {
    d.mu.Lock()
    defer d.mu.Unlock()

	if len(d.store) == 0 {
		return
	}

    now := time.Now()
    for k, itm := range d.store {
        if !itm.expiresAt.IsZero() && itm.expiresAt.Before(now) {
            delete(d.store, k)
        }
    }
	log.Println("cleaned up expired keys in the background")
}


