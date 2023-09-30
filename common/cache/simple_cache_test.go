package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"stock/common/log"

	"github.com/stretchr/testify/assert"
)

func TestSimpleCache(t *testing.T) {
	a := assert.New(t)
	sc := NewSimpleCache()
	sc.set("hello", "world", time.Now().Add(time.Hour))
	ret, ok := sc.get("hello")
	if !ok || ret.(string) != "world" {
		t.Fatal(ret)
	}
	sc.set("hello2", "world2", time.Now().Add(time.Second))
	time.Sleep(2 * time.Second)
	ret, ok = sc.get("hello2")
	if ok {
		t.Fatal(ret)
	}

	f := func() (interface{}, error) {
		return "world3", nil
	}

	_, err := sc.GetAndLoad("hello3", f, time.Hour)
	a.NoError(err)
	ret, ok = sc.get("hello3")
	if !ok || ret.(string) != "world3" {
		t.Fatal(ret)
	}

	sc.LastClearTime = time.Now().Add(time.Hour * -1)
	sc.Set("hehe", "haha", time.Hour)
	v, ok := sc.Get("hehe")
	a.True(ok)
	a.Equal(v.(string), "haha")
}

func TestGetSet(t *testing.T) {
	sc := NewSimpleCache()
	key := "test"
	_, ok := sc.get(key)
	assert.Equal(t, ok, false)
	sc.Set(key, 1, time.Second)
	value, ok := sc.get(key)
	assert.Equal(t, ok, true)
	assert.Equal(t, value.(int), 1)

	time.Sleep(time.Second * 2)
	_, ok = sc.get(key)
	assert.Equal(t, ok, false)
}

func TestSimpleCacheNil(t *testing.T) {
	sc := NewSimpleCache()
	value, err := sc.GetAndLoad("hello", func() (interface{}, error) {
		return "", nil
	}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	if str, ok := value.(string); !ok {
		t.Fatal(str, ok, value)
	}

	value1, err := sc.GetAndLoad("hello2", func() (interface{}, error) {
		var b []string
		return b, nil
	}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	if str1, ok := value1.([]string); !ok {
		t.Fatal(str1, ok, value1, "hehe")
	}

}

func TestSimpleCacheStress(t *testing.T) {
	a := assert.New(t)
	sc := NewSimpleCache()
	fmt.Println("Test TestSimpleCacheStress start ...")
	var wg sync.WaitGroup
	calledCount := 0
	for i := 0; i < 10000000; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := sc.GetAndLoad("cache-key", func() (interface{}, error) {
				time.Sleep(time.Second)
				calledCount++
				return "result", nil
			}, time.Second*5)
			a.NoError(err)
		}(i)
	}
	wg.Wait()
	fmt.Println("called load func count ", calledCount)
	fmt.Println("Test TestSimpleCacheStress end ...")
}

func TestSimpleCache_GetAndRefresh(t *testing.T) {
	sc := NewSimpleCache()
	fmt.Println("Test TestSimpleCacheStress start ...")
	var wg sync.WaitGroup
	calledCount := 0
	now := time.Now()
	for i := 0; i < 10000000; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			data, err := sc.GetOrSetRefresh("cache-key", func() (interface{}, error) {
				calledCount++
				log.Info("%d", calledCount)
				return calledCount, nil
			}, time.Second*50, time.Second)
			if err != nil {
				t.Error(err)
				return
			}

			secs, ok := data.(int)
			if !ok {
				t.Error(data)
				return
			}

			interval := time.Since(now).Seconds() - float64(secs)
			if interval > 2 || interval < -1 {
				t.Error(interval)
				return
			}
		}(i)
	}
	wg.Wait()

	fmt.Println("called load func count ", calledCount)
	fmt.Println("Test TestSimpleCacheStress end ...")
}
