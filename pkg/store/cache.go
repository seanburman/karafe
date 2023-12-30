package store

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"
)

type (
	cache[T interface{}] struct {
		createdAt  time.Time
		raw        *raw[T]
		serializer *serializer[T]
	}
	raw[T interface{}] struct {
		mu      sync.Mutex
		caches  map[CacheKey]*Item[T]
		history map[time.Time]map[CacheKey]Item[T]
		feed    chan map[time.Time]map[CacheKey]Item[T]
	}
	serializer[T interface{}] struct {
		mu        sync.Mutex
		serialize *func(SerializerConfig[T]) []byte
		feed      chan []byte
		history   map[time.Time][]byte
	}
	Item[T interface{}] struct {
		CreatedAt time.Time `json:"created_at"`
		Data      *T        `json:"data"`
	}
	cacheTimeoutConfig[T interface{}] struct {
		data       *T
		key        interface{}
		timeoutFun func(data *T)
		timeout    time.Duration
	}
	CacheKey                        interface{}
	SerializerConfig[T interface{}] struct {
		CreatedAt time.Time
		Data      []SerialData[T]
	}
	SerialData[T interface{}] struct {
		Key  CacheKey
		Item Item[T]
	}
)

// newCache[T] returns a new instance of *cache[T]
func newCache[T interface{}]() (data *cache[T]) {
	c := &cache[T]{
		createdAt: time.Now(),
		raw: &raw[T]{
			caches:  make(map[CacheKey]*Item[T]),
			history: make(map[time.Time]map[CacheKey]Item[T]),
			feed:    make(chan map[time.Time]map[CacheKey]Item[T], 1024),
		},
		serializer: &serializer[T]{
			feed:    make(chan []byte, 1024),
			history: make(map[time.Time][]byte),
		},
	}
	setup := make(chan bool)
	defer close(setup)
	go c.monitorChanges(setup)
	<-setup
	return c
}

func (c *cache[T]) monitorChanges(setup chan bool) {
	prev := c.copyRaw()
	setup <- true
	for {
		current := c.copyRaw()
		if !reflect.DeepEqual(current, prev) {
			c.cacheCopy(current)
			prev = current
		}
	}
}

func (c *cache[T]) copyRaw() map[CacheKey]Item[T] {
	c.raw.mu.Lock()
	defer c.raw.mu.Unlock()
	copy := make(map[CacheKey]Item[T])
	for k, v := range c.raw.caches {
		copy[k] = *v
	}
	return copy
}

func (c *cache[T]) cacheCopy(copy map[CacheKey]Item[T]) {
	c.raw.mu.Lock()
	t := time.Now()
	c.raw.history[t] = copy
	c.raw.feed <- c.raw.history
	c.raw.mu.Unlock()
	go c.serialize(t, copy)
}

func (c *cache[T]) SetSerializer(sf func(cfg SerializerConfig[T]) []byte) {
	c.serializer.mu.Lock()
	defer c.serializer.mu.Unlock()
	c.serializer.serialize = &sf
}

func (c *cache[T]) serialize(t time.Time, copy map[CacheKey]Item[T]) {
	if c.serializer.serialize == nil {
		return
	}

	config := SerializerConfig[T]{}
	config.CreatedAt = t
	for key, item := range copy {
		config.Data = append(config.Data, SerialData[T]{
			Key:  key,
			Item: item,
		})
	}

	c.serializer.mu.Lock()
	serialize := *c.serializer.serialize
	dataSerialized := serialize(config)
	c.serializer.history[config.CreatedAt] = dataSerialized
	c.serializer.feed <- dataSerialized
	c.serializer.mu.Unlock()
}

func (c *cache[T]) FeedRaw() chan map[time.Time]map[CacheKey]Item[T] {
	c.raw.mu.Lock()
	defer c.raw.mu.Unlock()
	return c.raw.feed
}

func (c *cache[T]) FeedSerialized() chan []byte {
	c.serializer.mu.Lock()
	defer c.serializer.mu.Unlock()
	return c.serializer.feed
}

func (c *cache[T]) HistoryRaw() map[time.Time]map[CacheKey]Item[T] {
	c.raw.mu.Unlock()
	defer c.raw.mu.Unlock()
	return c.raw.history
}

func (c *cache[T]) HistorySerialized() map[time.Time][]byte {
	c.serializer.mu.Lock()
	defer c.serializer.mu.Unlock()
	return c.serializer.history
}

func (c *cache[T]) Get(key CacheKey) (*Item[T], bool) {
	c.raw.mu.Lock()
	defer c.raw.mu.Unlock()

	data := c.raw.caches[key]
	if data == nil {
		return nil, false
	}
	return data, true
}

func (c *cache[T]) All() map[CacheKey]*Item[T] {
	c.raw.mu.Lock()
	defer c.raw.mu.Unlock()
	return c.raw.caches
}

func (c *cache[T]) Save(data *T, key CacheKey) error {
	c.raw.mu.Lock()
	defer c.raw.mu.Unlock()

	if c.raw.caches[key] != nil {
		return fmt.Errorf("duplicate cache key: %v", key)
	}
	c.raw.caches[key] = &Item[T]{
		Data:      data,
		CreatedAt: time.Now(),
	}
	return nil
}

func (c *cache[T]) SaveWithTimeout(cfg cacheTimeoutConfig[T]) error {
	c.Save(cfg.data, cfg.key)
	if !(cfg.timeout > time.Second*0) {
		return fmt.Errorf(
			"cache not set for timeout: %v; timeout must be greater than 0", cfg.timeout,
		)
	}

	go func() {
		timer := time.NewTimer(cfg.timeout)
		<-timer.C
		cache, ok := c.Get(cfg.key)
		if !ok {
			log.Panicf("could not get cache with key %v", cfg.key)
		}
		err := c.Delete(cfg.key)
		if err != nil {
			log.Panicf("%v", err)
		}
		cfg.timeoutFun(cache.Data)
	}()
	return nil
}

func (c *cache[T]) Delete(key interface{}) error {
	c.raw.mu.Lock()
	defer c.raw.mu.Unlock()

	if c.raw.caches[key] == nil {
		return fmt.Errorf("no cache with key: %v", key)
	}
	delete(c.raw.caches, key)
	return nil
}

func NewCacheTimeoutConfig[T interface{}](
	data *T,
	key interface{},
	timeoutFun func(data *T),
	timeout time.Duration,
) cacheTimeoutConfig[T] {
	return cacheTimeoutConfig[T]{
		data:       data,
		key:        key,
		timeoutFun: timeoutFun,
		timeout:    timeout,
	}
}
