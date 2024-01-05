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
		mu        sync.Mutex
		createdAt time.Time
		raw       *raw[T]
		reducer   *reducer[T]
	}
	raw[T interface{}] struct {
		caches  map[CacheKey]*Item[T]
		history map[time.Time]map[CacheKey]Item[T]
		feed    chan map[time.Time]map[CacheKey]Item[T]
	}
	reducer[T interface{}] struct {
		reduce  *func(ReducerConfig[T]) interface{}
		feed    chan interface{}
		history map[time.Time]interface{}
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
	CacheKey                     interface{}
	ReducerConfig[T interface{}] struct {
		CreatedAt time.Time
		Data      []ReducerData[T]
	}
	ReducerData[T interface{}] struct {
		Key  CacheKey
		Item Item[T]
	}
)

func newCache[T interface{}]() (data *cache[T]) {
	c := &cache[T]{
		createdAt: time.Now(),
		raw: &raw[T]{
			caches:  make(map[CacheKey]*Item[T]),
			history: make(map[time.Time]map[CacheKey]Item[T]),
			feed:    make(chan map[time.Time]map[CacheKey]Item[T], 1024),
		},
		reducer: &reducer[T]{
			feed:    make(chan interface{}, 1024),
			history: make(map[time.Time]interface{}),
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
		//TODO: Apply reducer to this data so we can see changes the user
		// is looking for.
		current := c.copyRaw()
		if !reflect.DeepEqual(current, prev) {
			c.cacheCopy(current)
			prev = current
		}
	}
}

func (c *cache[T]) copyRaw() map[CacheKey]Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	copy := make(map[CacheKey]Item[T])
	for k, v := range c.raw.caches {
		copy[k] = *v
	}
	return copy
}

func (c *cache[T]) cacheCopy(copy map[CacheKey]Item[T]) {
	c.mu.Lock()
	t := time.Now()
	c.raw.history[t] = copy
	c.raw.feed <- c.raw.history
	c.mu.Unlock()
	go c.reduce(t, copy)
}

func (c *cache[T]) SetReducer(sf func(cfg ReducerConfig[T]) interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reducer.reduce = &sf
}

func (c *cache[T]) reduce(t time.Time, copy map[CacheKey]Item[T]) {
	if c.reducer.reduce == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	config := ReducerConfig[T]{}
	config.CreatedAt = t
	for key, item := range copy {
		config.Data = append(config.Data, ReducerData[T]{
			Key:  key,
			Item: item,
		})
	}

	reduce := *c.reducer.reduce
	dataReduced := reduce(config)
	c.reducer.history[config.CreatedAt] = dataReduced
	c.reducer.feed <- dataReduced
}

func (c *cache[T]) RawFeed() chan map[time.Time]map[CacheKey]Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.raw.feed
}

func (c *cache[T]) ReducerFeed() chan interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reducer.feed
}

func (c *cache[T]) RawHistory() map[time.Time]map[CacheKey]Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.raw.history
}

func (c *cache[T]) ReducerHistory() map[time.Time]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reducer.history
}

func (c *cache[T]) GetOne(key CacheKey) (*Item[T], bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := c.raw.caches[key]
	if data == nil {
		return nil, false
	}
	return data, true
}

func (c *cache[T]) GetAll() map[CacheKey]*Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.raw.caches
}

func (c *cache[T]) Save(data *T, key CacheKey) error {
	c.mu.Lock()
	defer c.mu.Unlock()

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
		cache, ok := c.GetOne(cfg.key)
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
	c.mu.Lock()
	defer c.mu.Unlock()

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
