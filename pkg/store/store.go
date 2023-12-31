// Package store blah blah blah
package store

import (
	"fmt"
	"log"
	"sync"

	"github.com/seanburman/kaw/cmd/api"
)

const serverStoreCache StoreKey = "server_store_cache"

var serverStore = NewStore("server_store")

func init() {
	_, err := CreateStoreCache[api.Server](serverStore, serverStoreCache)
	if err != nil {
		log.Fatal(err)
	}
}

type (
	Store struct {
		mu   sync.Mutex
		name string
		data map[interface{}]interface{}
		keys []StoreKey
	}
	StoreKey string
)

func NewStore(name string) *Store {
	return &Store{
		name: name,
		data: make(map[interface{}]interface{}),
	}
}

func (s *Store) Serve(port string, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cache, err := UseStoreCache[api.Server](serverStore, serverStoreCache)
	if err != nil {
		log.Fatal(err)
	}
	for _, cs := range cache.All() {
		if cs.Data.Config.Port == port {
			return fmt.Errorf("port %v already taken with path %v", port, cs.Data.Config.Path)
		}
	}

	server, err := api.NewServer(api.ServerConfig{
		Port: port,
		Path: path,
	})
	if err != nil {
		return err
	}

	if err = cache.Save(server, s.name); err != nil {
		return err
	}
	return nil
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
