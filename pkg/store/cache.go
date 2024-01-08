package store

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"
)

type (
	cache[T any] struct {
		mu        sync.Mutex
		createdAt time.Time
		raw       *raw[T]
		reducer   *reducer[T]
	}
	raw[T any] struct {
		caches  map[CacheKey]*Item[T]
		history map[time.Time]map[CacheKey]Item[T]
		feed    chan map[time.Time]map[CacheKey]Item[T]
	}
	reducer[T any] struct {
		reduce  *func([]ReducerData[T]) []any
		feed    chan any
		history map[time.Time]any
	}
	Item[T any] struct {
		CreatedAt time.Time `json:"created_at"`
		Data      *T        `json:"data"`
	}
	cacheTimeoutConfig[T any] struct {
		data       *T
		key        any
		timeoutFun func(data *T)
		timeout    time.Duration
	}
	CacheKey                  any
	ReducerFunc[T any, U any] func(previous []U, current T) (next []U)
	ReducerConfig[T any]      struct {
		CreatedAt time.Time
		Data      []ReducerData[T]
	}
	ReducerData[T any] struct {
		Key       CacheKey  `json:"key"`
		CreatedAt time.Time `json:"created_at"`
		Data      T         `json:"data"`
	}
)

func newCache[T any]() (data *cache[T]) {
	c := &cache[T]{
		createdAt: time.Now(),
		raw: &raw[T]{
			caches:  make(map[CacheKey]*Item[T]),
			history: make(map[time.Time]map[CacheKey]Item[T]),
			feed:    make(chan map[time.Time]map[CacheKey]Item[T], 1024),
		},
		reducer: &reducer[T]{
			feed:    make(chan any, 1024),
			history: make(map[time.Time]any),
		},
	}
	return c
}

func (c *cache[T]) monitorChanges(setup chan bool) {
	pcopy := c.copyRaw()
	setup <- true
	prev := c.reduce(pcopy)
	c.cacheCopy(pcopy)
	for {
		ccopy := c.copyRaw()
		current := c.reduce(ccopy)
		if current == nil {
			continue
		}
		// Check for changes
		if !reflect.DeepEqual(prev, current) {
			c.cacheCopy(ccopy)
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
	go c.cacheReduction(t, c.reduce(copy))
}

func newCacheReducer[T, U any](f ReducerFunc[T, U]) func([]ReducerData[T]) []U {
	return func(rd []ReducerData[T]) []U {
		var u []U
		for _, d := range rd {
			u = f(u, d.Data)
		}
		return u
	}
}

func (c *cache[T]) SetReducer(rf ReducerFunc[T, any]) {
	cr := newCacheReducer(rf)
	c.mu.Lock()
	c.reducer.reduce = &cr
	c.mu.Unlock()

	setup := make(chan bool)
	defer close(setup)
	go c.monitorChanges(setup)
	<-setup
}

func (c *cache[T]) DefaultReducer(previous []any, current T) (next []any) {
	return append(previous, current)
}

func (c *cache[T]) reduce(copy map[CacheKey]Item[T]) []any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.reducer.reduce == nil {
		log.Printf("no reducer set for cache type: %v, setting DefaultReducer", reflect.TypeOf(*new(T)))
		c.SetReducer(c.DefaultReducer)
		return nil
	}

	data := []ReducerData[T]{}
	for key, item := range copy {
		data = append(data, ReducerData[T]{
			Key:       key,
			CreatedAt: item.CreatedAt,
			Data:      *item.Data,
		})
	}
	reduce := *c.reducer.reduce
	return reduce(data)
}

func (c *cache[T]) cacheReduction(t time.Time, r []any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reducer.history[t] = r
	c.reducer.feed <- r
}

func (c *cache[T]) RawFeed() chan map[time.Time]map[CacheKey]Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.raw.feed
}

func (c *cache[T]) ReducerFeed() chan any {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reducer.feed
}

func (c *cache[T]) RawHistory() map[time.Time]map[CacheKey]Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.raw.history
}

func (c *cache[T]) ReducerHistory() map[time.Time]any {
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

func (c *cache[T]) Cache(data *T, key CacheKey) error {
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

func (c *cache[T]) CacheWithTimeout(cfg cacheTimeoutConfig[T]) error {
	c.Cache(cfg.data, cfg.key)
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
			log.Panicln(err)
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

func NewCacheTimeoutConfig[T any](
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
