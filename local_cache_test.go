package local_cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var dataSource sync.Map

func initDataSource() {
	dataSource = sync.Map{}
	dataSource.Store("test", "test")
	dataSource.Store("value", "value")
}
func updateDataSource() {
	dataSource.Store("test", "test2")
	dataSource.Store("value", "value2")
}
func timerLoadFunc(key string) (interface{}, error) {
	val, _ := dataSource.Load(key)
	fmt.Printf("load key:%+v \n", key)
	return val, nil
}

func TestTimer(t *testing.T) {
	cache := new(TimerLocalCache)
	initDataSource()
	cache.Init(timerLoadFunc, DefaultErrorFunction, 2*time.Second, "test", "value")
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Printf("update:dataSource\n")
		updateDataSource()
	}()
	for {
		testVal, _ := cache.GetCache("test")
		fmt.Printf("TestLoadKey: %+v,value:%+v\n", "test", testVal)
		valueVal, _ := cache.GetCache("value")
		fmt.Printf("TestLoadKey: %+v,value:%+v\n", "value", valueVal)
		time.Sleep(50 * time.Millisecond)
	}
}
