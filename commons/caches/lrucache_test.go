package caches

import (
	"sync"
	"testing"
)

func TestReadTimeoutCleanupConcurrentAccess(t *testing.T) {
	cache := NewLruCache("concurrent-cleanup", 16, nil)
	cache.Add("active", "value")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			cache.Add("active", "value")
			cache.Get("active")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			cache.cleanOldestByReadTime(0)
		}
	}()
	wg.Wait()
	if value, ok := cache.Get("active"); !ok || value != "value" {
		t.Fatal("cleanup removed an unexpired entry")
	}
	cache.SetBatchEvict(1, func(items []CacheItem) {
		// Batch callbacks must remain outside the cache lock.
		cache.Len()
		if len(items) != 1 || items[0].Key != "active" {
			t.Errorf("unexpected evicted items: %+v", items)
		}
	})
	cache.cleanOldestByReadTime(1 << 62)
	if cache.Len() != 0 {
		t.Fatal("cleanup retained an expired entry")
	}
}
