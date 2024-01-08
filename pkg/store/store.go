// Package store blah blah blah
package store

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/labstack/gommon/log"
	"github.com/seanburman/kachekrow/cmd/api"
	"github.com/seanburman/kachekrow/cmd/gui"
	"github.com/seanburman/kachekrow/pkg/connection"
)

const ServerStore StoreKey = "servers"

var ServerCache = CacheKey("servers_cache")

var storeManager struct {
	mu     sync.Mutex
	stores map[StoreKey]*Store
} = struct {
	mu     sync.Mutex
	stores map[StoreKey]*Store
}{
	stores: make(map[StoreKey]*Store),
}

func init() {
	serverStore, _ := NewStore(ServerStore)
	_, err := NewCache[api.Server](serverStore, ServerCache)
	if err != nil {
		log.Fatal(err)
	}

	port := ":" + os.Getenv("KACHE_KROW_PORT")
	if port == ":" {
		port = ":8080"
	}
	serverStore.Serve(port, "/store")
}

type (
	Store struct {
		mu   sync.Mutex
		key  StoreKey
		data map[CacheKey]any
	}
	StoreKey string
)

func KacheKrow() {
	fmt.Println(`𓅩 KACHE KROW`)
	gui.ListenCommands()
}

func UseStore(key StoreKey) *Store {
	storeManager.mu.Lock()
	defer storeManager.mu.Unlock()
	store, ok := storeManager.stores[key]
	if !ok {
		return nil
	}
	return store
}

func NewStore(key StoreKey) (*Store, error) {
	storeManager.mu.Lock()
	defer storeManager.mu.Unlock()
	s := &Store{
		key:  key,
		data: make(map[CacheKey]any),
	}
	_, ok := storeManager.stores[key]
	if ok {
		return nil, fmt.Errorf("store with key '%v' already exists", key)
	}
	storeManager.stores[key] = s
	return s, nil
}

func (s *Store) Serve(port string, path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("empty store path for port: %s", port)
	}

	cfg := api.NewConfig(port, path, string(s.key))
	server, err := api.NewServer(cfg)
	if err != nil {
		return err
	}

	caches, err := UseCache[api.Server](ServerStore, ServerCache)
	if err != nil {
		log.Fatal(err)
	}

	caches.SetReducer(func(previous []any, current api.Server) (next []any) {
		return append(previous, current.Config())
	})

	if err = caches.Cache(server, s.key); err != nil {
		return err
	}

	server.SetOnNewConnection(func(c *connection.Connection) {
		c.Publish(caches.ReducerHistory())
	})

	server.ListenAndServe()
	return nil
}

func NewCache[Cache any](s *Store, key CacheKey) (*cache[Cache], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	if ok {
		return nil, fmt.Errorf("key '%v' already exists", key)
	}

	cs := newCache[Cache]()
	s.data[key] = cs

	return cs, nil
}

func UseCache[Cache any](s StoreKey, c CacheKey) (*cache[Cache], error) {
	store := UseStore(s)
	if store == nil {
		return nil, fmt.Errorf("no store with key '%v'", s)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	data, ok := store.data[c]
	if !ok {
		return nil, fmt.Errorf("no cache with key '%v'", c)
	}
	cache, ok := data.(*cache[Cache])
	if !ok {
		return nil, fmt.Errorf("invalid type for cache with key %v", c)
	}
	return cache, nil
}
