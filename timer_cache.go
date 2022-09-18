package local_cache

import (
	"sync"
	"time"
)

// TimerLocalCache init keys with a single goroutine
// AddCache must before Init
type TimerLocalCache struct {
	m             sync.Map
	flushDuration time.Duration
	keys          []string
}

func (t *TimerLocalCache) Init(f LoadFunction, ef ErrorFunction, duration time.Duration, keys ...string) {
	t.m = sync.Map{}
	t.flushDuration = duration
	t.AddCache(keys...)
	err := t.run(f, ef, t.keys...)
	if err != nil {
		panic(err)
	}
}
func (t *TimerLocalCache) AddCache(keys ...string) []error {
	t.keys = append(t.keys, keys...)
	return nil
}
func (t *TimerLocalCache) run(f LoadFunction, ef ErrorFunction, keys ...string) []error {
	var healthKeys []string
	var errs []error
	for _, key := range keys {
		val, err := f(key)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		healthKeys = append(healthKeys, key)
		t.m.Store(key, val)
	}
	go func() {
		for ; ; {
			for _, key := range healthKeys {
				val, err := f(key)
				if err != nil {
					ef(key, err)
				}
				t.m.Store(key, val)
			}
			time.Sleep(t.flushDuration)
		}

	}()
	return errs
}

func (t *TimerLocalCache) GetCache(key string) (interface{}, bool) {
	return t.m.Load(key)
}
