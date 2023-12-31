// Package store blah blah blah
package store

import (
	"fmt"
	"os/exec"
	"sync"

	"github.com/labstack/gommon/log"
	"github.com/seanburman/kaw/cmd/api"
	"github.com/seanburman/kaw/cmd/gui"
	"github.com/seanburman/kaw/handlers"
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

	serverStore.Serve(":8080", "/store")
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

func Kaw(web bool, desktop bool) {
	if desktop {
		fmt.Println(`𓅩 KAW! Opening GUI`)
		cmnd := exec.Command("cmd/kaw")
		cmnd.Start()
	}
	if web {
		gui.OpenURL("http://localhost:8080/store")
	}
	go gui.ListenCommands()
}

func NewStore(name string) *Store {
	return &Store{
		name: name,
		data: make(map[interface{}]interface{}),
	}
}

func (s *Store) Serve(port string, path string) error {

	caches, err := UseStoreCache[api.Server](serverStore, serverStoreCache)
	if err != nil {
		log.Fatal(err)
	}
	for _, cache := range caches.All() {
		if cache.Data.Config.Port == port {
			return fmt.Errorf("port %v already taken with path %v", port, cache.Data.Config.Path)
		}
	}

	server, err := api.NewServer(api.ServerConfig{
		Port: port,
		Path: path,
	})
	if err != nil {
		return err
	}

	wsRoute := api.NewRoute("/ws")
	wsRoute.HandleFunc("", handlers.HandleGetWebSocket)

	server.UseRoute(wsRoute)

	if err = caches.Save(server, s.name); err != nil {
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
