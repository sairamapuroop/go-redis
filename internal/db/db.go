package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type item struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

type DB struct {
	mu    sync.RWMutex
	store map[string]item
	dirty bool
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

	if itm.ExpiresAt.IsZero() || itm.ExpiresAt.After(time.Now()) {
		return itm.Value, true
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
		Value:     val,
		ExpiresAt: expiresAt,
	}

	d.dirty = true
}

func (d *DB) Delete(key string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	_, ok := d.store[key]

	if ok {
		delete(d.store, key)
		d.dirty = true
	}


	return ok

}

func (d *DB) Flush() {
	d.mu.Lock()
	d.store = make(map[string]item)
	d.mu.Unlock()

	d.dirty = true
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
	deleted := 0
	for k, itm := range d.store {
		if !itm.ExpiresAt.IsZero() && itm.ExpiresAt.Before(now) {
			delete(d.store, k)
			deleted++

		}
	}

	if deleted > 0 {
		log.Println("cleaned up expired keys in the background")
	}
}

// persistence using snapshot approach

func (d *DB) Save(filename string) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	fmt.Println("Saving", len(d.store), "items to", filename)
	encoder := json.NewEncoder(file)
	return encoder.Encode(d.store)

}

func (d *DB) Load(filename string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	defer file.Close()

	data := make(map[string]item)
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	//Remove expired keys
	now := time.Now()
	for k, v := range data {
		if !v.ExpiresAt.IsZero() && v.ExpiresAt.Before(now) {
			delete(data, k)
		}
	}

	d.store = data
	return nil

}

func (d *DB) StartPersistence(interval time.Duration, filename string, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				
				d.mu.Lock()
				if !d.dirty {
					d.mu.Unlock()
					continue
				}
				d.dirty = false
				d.mu.Unlock()

				if err := d.Save(filename); err != nil {
					log.Printf("error saving snapshot: %v", err)
				}
				log.Println("snapshot saved")
			case <-stopCh:
				return
			}
		}
	}()
}
