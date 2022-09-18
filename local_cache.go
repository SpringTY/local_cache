package local_cache

import (
	"log"
	"time"
)

type LocalCache interface {
	Init(globalLoadFunc LoadFunction, globalErrorFunc ErrorFunction, flushDuration time.Duration, keys ...string)
	AddCache(f LoadFunction, ef ErrorFunction, keys ...string) []error
	GetCache(key string) (interface{}, bool)
}
type LoadFunction func(key string) (interface{}, error)
type ErrorFunction func(string, error)

func DefaultErrorFunction(key string, err error) {
	log.Printf("[Error] load key: %s, err: %+v", key, err)
}
