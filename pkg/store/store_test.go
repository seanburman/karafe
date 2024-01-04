package store

import (
	"fmt"
	"reflect"
	"testing"
)

func Example() {
	fmt.Println("this is an example")
}

func TestNewStore(t *testing.T) {
	store := NewStore("test")
	storeType := reflect.TypeOf(store)
	expected := "*store.Store"
	actual := storeType.String()
	if actual != expected {
		t.Errorf("expected NewStore to be of type %v; got %v", expected, actual)
	}

	if store.data == nil {
		t.Errorf("expected data map to be instantiated")
	}
}

func TestKeys(t *testing.T) {
	store := NewStore("test")
	key := "test"
	_, err := CreateStoreCache[int](store, StoreKey(key))
	if err != nil {
		t.Error(err)
	}
	keys := store.Keys()
	if keys[0] != "test" {
		t.Error("expected keys to contain new CacheKey")
	}
}

func TestCreateStoreCache(t *testing.T) {
	store := NewStore("test")
	key := 0
	_, err := CreateStoreCache[int](store, StoreKey(key))
	if err != nil {
		t.Error(err)
	}
	_, err = CreateStoreCache[int](store, StoreKey(key))
	if err == nil {
		t.Error("expected duplicate key error")
	}
}

func TestUseStoreCache(t *testing.T) {
	store := NewStore("test")
	key := 0
	cache, err := CreateStoreCache[int](store, StoreKey(key))
	if err != nil {
		t.Error(err)
	}
	data := 123
	cache.Save(&data, key)
	cache, err = UseStoreCache[int](store, StoreKey(key))
	if err != nil {
		t.Error(err)
	}
	if len(cache.GetAll()) != 1 {
		t.Error("cache not retrieved")
	}

	_, err = UseStoreCache[int](store, StoreKey(2))
	if err == nil {
		t.Error("expected no data with key error")
	}

	_, err = UseStoreCache[string](store, StoreKey(0))
	if err == nil {
		t.Error("expected invalid type for cache with key error")
	}
}
