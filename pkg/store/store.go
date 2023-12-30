// Package store blah blah blah
package store

import (
	"fmt"
	"sync"
)

type (
	Store struct {
		mu   sync.Mutex
		data map[interface{}]interface{}
		keys []StoreKey
	}
	StoreKey int
)

func NewStore() *Store {
	s := &Store{
		data: make(map[interface{}]interface{}),
	}
	return s
}

func (s *Store) Keys() []StoreKey {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.keys
}

func CreateStoreCache[Cache interface{}](s *Store, key StoreKey) (*cache[Cache], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range s.keys {
		if v == key {
			return nil, fmt.Errorf("key %v already exists", key)
		}
	}
	s.keys = append(s.keys, key)
	cs := newCache[Cache]()
	s.data[key] = cs

	return cs, nil
}

func UseStoreCache[Cache interface{}](s *Store, key StoreKey) (*cache[Cache], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("no data with key %v", key)
	}
	cache, ok := data.(*cache[Cache])
	if !ok {
		return nil, fmt.Errorf("invalid type for cache with key %v", key)
	}
	return cache, nil
}
