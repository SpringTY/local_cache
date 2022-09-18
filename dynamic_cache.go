package local_cache

import (
	"github.com/cespare/xxhash"
	"math"
	"sync"
	"time"
)

const DefaultDynamicLocalCachePartition = 4

// DynamicLocalCache every key with a single goroutine to flush
// can load key with runtime, next time will support many key with a single goroutine
// Init must before to AddCache
type DynamicLocalCache struct {
	m               sync.Map //cache Map
	globalLoadFunc  LoadFunction
	globalErrorFunc ErrorFunction
	locks           []sync.Locker
	lockerSize      int64
	flushDuration   time.Duration
}

func (t *DynamicLocalCache) WithLockerSize(lockerSize int64) *DynamicLocalCache {
	t.lockerSize = lockerSize
	return t
}
func (t *DynamicLocalCache) Init(globalLoadFunc LoadFunction, globalErrorFunc ErrorFunction, flushDuration time.Duration, keys ...string) {
	if globalErrorFunc == nil {
		globalErrorFunc = DefaultErrorFunction
	}
	t.m = sync.Map{}
	t.flushDuration = flushDuration
	t.globalLoadFunc = globalLoadFunc
	t.globalErrorFunc = globalErrorFunc
	var lockers []sync.Locker
	if t.lockerSize == 0 {
		t.lockerSize = int64(math.Max(DefaultDynamicLocalCachePartition, float64(len(keys))))
	}
	for i := int64(0); i < t.lockerSize; i++ {
		lockers = append(lockers, &sync.Mutex{})
	}
	err := t.AddCache(nil, nil, keys...)
	if len(err) != 0 {
		panic(err)
	}
}

func (t *DynamicLocalCache) AddCache(f LoadFunction, ef ErrorFunction, keys ...string) []error {

	if ef == nil {
		ef = t.globalErrorFunc
	}
	if f == nil {
		f = t.globalLoadFunc
	}
	var errs []error
	for _, key := range keys {
		err := t.addCache(f, ef, key)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
func (t *DynamicLocalCache) addCache(f LoadFunction, ef ErrorFunction, key string) error {

	if _, loaded := t.m.Load(key); !loaded {
		// firstly getValue
		value, err := f(key)
		if err != nil {
			return err
		}
		t.m.Store(key, value)
		// secondly try to run flush goroutine
		locker := t.GetKeyLock(key)
		defer locker.Unlock()
		locker.Lock()
		if _, loaded := t.m.Load(key); !loaded {
			// double check to make sure : only one goroutine load a single key
			go func() {
				for {
					value, err := f(key)
					if err != nil {
						ef(key, err)
					}
					t.m.Store(key, value)
					time.Sleep(t.flushDuration)
				}
			}()
		}
	}
	return nil
}
func (t *DynamicLocalCache) GetCache(key string) (interface{}, bool) {
	return t.m.Load(key)
}
func (t *DynamicLocalCache) GetKeyLock(key string) sync.Locker {
	// diff key with diff lock
	lockerIndex := xxhash.Sum64String(key) % uint64(t.lockerSize)
	return t.locks[lockerIndex]
}
