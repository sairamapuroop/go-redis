package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type ValueType string

const (
	StringType ValueType = "string"
	ListType   ValueType = "list"
	SetType    ValueType = "set"
	HashType   ValueType = "hash"
)

type item struct {
	Type        ValueType           `json:"type"`
	StringValue string              `json:"string_value,omitempty"`
	ListValue   []string            `json:"list_value,omitempty"`
	SetValue    map[string]struct{} `json:"set_value,omitempty"`
	HashValue   map[string]string   `json:"hash_value,omitempty"`
	ExpiresAt   time.Time           `json:"expires_at"`
}

type DB struct {
	mu    sync.RWMutex
	store map[string]*item
	dirty bool
}

func New() *DB {
	return &DB{
		store: make(map[string]*item),
	}
}

func (d *DB) Get(key string) (string, bool) {
	d.mu.RLock()
	itm, ok := d.store[key]
	d.mu.RUnlock()

	if !ok || itm.Type != StringType {
		return "", false
	}

	if itm.ExpiresAt.IsZero() || itm.ExpiresAt.After(time.Now()) {
		return itm.StringValue, true
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

	d.store[key] = &item{
		Type:        StringType,
		StringValue: val,
		ExpiresAt:   expiresAt,
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
	d.store = make(map[string]*item)
	d.mu.Unlock()

	d.dirty = true
}

// List Datastructure
func (d *DB) LPush(key string, values ...string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	itm, exists := d.store[key]

	if !exists {
		itm = &item{Type: ListType}
	}

	if itm.Type != ListType {
		return 0
	}

	itm.ListValue = append(values, itm.ListValue...)

	d.store[key] = itm

	d.dirty = true

	return len(itm.ListValue)

}

func (d *DB) RPush(key string, values ...string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	itm, exists := d.store[key]
	if !exists {
		itm = &item{Type: ListType}
	}

	if itm.Type != ListType {
		return 0
	}

	itm.ListValue = append(itm.ListValue, values...)
	d.store[key] = itm

	d.dirty = true

	return len(itm.ListValue)

}

func (d *DB) LRange(key string, start, end int) []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	itm, ok := d.store[key]
	if !ok || itm.Type != ListType {
		return nil
	}

	l := len(itm.ListValue)
	if start < 0 {
		start = l + start
		if start < 0 {
			start = 0
		}
	}

	if end < 0 {
		end = l + end
	}

	if end >= l {
		end = l - 1
	}

	if start > end {
		return nil
	}

	log.Println(start, end+1)

	return itm.ListValue[start : end+1]

}

func (d *DB) SAdd(key string, members ...string) {
    d.mu.Lock()
    defer d.mu.Unlock()

    itm, exists := d.store[key]
    if !exists {
        itm = &item{Type: SetType, SetValue: make(map[string]struct{})}
    }
    if itm.Type != SetType {
        return
    }

    for _, m := range members {
        itm.SetValue[m] = struct{}{}
    }

    d.store[key] = itm
    d.dirty = true
}

func (d *DB) SMembers(key string) []string {
    d.mu.RLock()
    defer d.mu.RUnlock()

    itm, ok := d.store[key]
    if !ok || itm.Type != SetType {
        return nil
    }

    members := make([]string, 0, len(itm.SetValue))
    for k := range itm.SetValue {
        members = append(members, k)
    }
    return members
}

func (d *DB) HSet(key, field, value string) {
    d.mu.Lock()
    defer d.mu.Unlock()

    itm, exists := d.store[key]
    if !exists {
        itm = &item{Type: HashType, HashValue: make(map[string]string)}
    }
    if itm.Type != HashType {
        return
    }

    itm.HashValue[field] = value
    d.store[key] = itm
    d.dirty = true
}

func (d *DB) HGet(key, field string) (string, bool) {
    d.mu.RLock()
    defer d.mu.RUnlock()

    itm, ok := d.store[key]
    if !ok || itm.Type != HashType {
        return "", false
    }

    val, exists := itm.HashValue[field]
    return val, exists
}

func (d *DB) HGetAll(key string) []string {
    d.mu.RLock()
    defer d.mu.RUnlock()

    itm, ok := d.store[key]
    if !ok || itm.Type != HashType {
        return nil
    }

    result := make(map[string]string, len(itm.HashValue))
    for k, v := range itm.HashValue {
        result[k] = v
    }

	resultstring := make([]string, 0, len(result)*2)

	for k,v := range result {
		resultstring = append(resultstring, k)
		resultstring = append(resultstring, v)
	}

    return resultstring
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

	data := make(map[string]*item)
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
