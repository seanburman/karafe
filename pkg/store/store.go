// Package store blah blah blah
package store

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/labstack/gommon/log"
	"github.com/seanburman/kaw/cmd/api"
	"github.com/seanburman/kaw/cmd/gui"
)

const serverStoreCache StoreKey = "server_store_cache"

var serverStore = NewStore("server_store")

func init() {
	cache, err := CreateStoreCache[api.Server](serverStore, serverStoreCache)
	if err != nil {
		log.Fatal(err)
	}
	cache.SetSerializer(func(cfg SerializerConfig[api.Server]) []byte {
		for _, server := range cfg.Data {
			fmt.Printf("Caching store %s\n", server.Key)
		}
		return []byte{}
	})

	port := os.Getenv("KAW_PORT")
	if port == "" {
		port = ":8080"
	}
	serverStore.Serve(port, "/store")
}

type (
	Store struct {
		mu   sync.Mutex
		key  string
		data map[interface{}]interface{}
		keys []StoreKey
	}
	StoreKey string
)

func Kaw() {
	fmt.Println(`𓅩 KAW! Kaching At Will`)
	gui.ListenCommands()
}

func NewStore(key string) *Store {
	return &Store{
		key:  key,
		data: make(map[interface{}]interface{}),
	}
}

func (s *Store) Serve(port string, path string) error {
	path = strings.TrimSpace(path)

	caches, err := UseStoreCache[api.Server](serverStore, serverStoreCache)
	if err != nil {
		log.Fatal(err)
	}
	for _, cache := range caches.All() {
		if cache.Data.Config.Port == port {
			return fmt.Errorf("store port %v already taken with path %v", port, cache.Data.Config.Path)
		}
	}

	server, err := api.NewServer(api.ServerConfig{
		Name: s.key,
		Port: port,
		Path: path,
	})
	if err != nil {
		return err
	}

	if err = caches.Save(server, s.key); err != nil {
		return err
	}

	server.ListenAndServe()
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
