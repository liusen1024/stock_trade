package cache

import (
	"errors"
	"sync"
	"time"

	"stock/common/log"

	"golang.org/x/sync/singleflight"
)

// SimpleCache  简单的一个缓存,没有LRU功能, 每一个小时会自动清除无人访问的过期的键值
type SimpleCache struct {
	data          map[string]interface{}
	expireTime    map[string]time.Time        //这里的expire指的是key的expire，而不是value的expire
	refreshChan   map[string]chan interface{} //通知刷新函数退出的信号通道
	LastClearTime time.Time                   //上一次清除cache的时间
	lock          *sync.RWMutex
	sfOnce        singleflight.Group
}

// NewSimpleCache 创建
func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		data:          make(map[string]interface{}),
		expireTime:    make(map[string]time.Time),
		lock:          &sync.RWMutex{},
		sfOnce:        singleflight.Group{},
		refreshChan:   make(map[string]chan interface{}),
		LastClearTime: time.Now(),
	}
}

// Set 第二个域表示是否存在这个key
func (sc *SimpleCache) Set(key string, val interface{}, expireDuration time.Duration) {
	sc.set(key, val, time.Now().Add(expireDuration))
}

// set 设置值和过期时间
func (sc *SimpleCache) set(key string, val interface{}, expireTime time.Time) {
	sc.lock.Lock()
	defer sc.lock.Unlock()
	sc.data[key] = val
	sc.expireTime[key] = expireTime
	//这里每10分钟会定期清空一下无人访问的并且过期的key
	sc.clearExpire(time.Now())
}

func (sc *SimpleCache) clearExpire(now time.Time) {
	if now.Sub(sc.LastClearTime).Minutes() < 10 {
		return
	}

	log.Infof("start clearExpire, keys %d", len(sc.data))
	expireKeys := make([]string, 0)
	for key, expireTime := range sc.expireTime {
		if expireTime.Before(now) {
			expireKeys = append(expireKeys, key)
		}
	}
	for _, key := range expireKeys {
		delete(sc.data, key)
		delete(sc.expireTime, key)
		if ch, ok := sc.refreshChan[key]; ok {
			close(ch)
			delete(sc.refreshChan, key)
		}

	}
	sc.LastClearTime = now
	log.Infof("after clearExpire, keys %d", len(sc.data))
}

// Get 第二个域表示是否存在这个key
func (sc *SimpleCache) Get(key string) (interface{}, bool) {
	return sc.get(key)
}

// get 获取缓存结果，false表示没有拿到值
func (sc *SimpleCache) get(key string) (interface{}, bool) {
	sc.lock.RLock()
	defer sc.lock.RUnlock()
	val, ok := sc.data[key]
	if !ok {
		return nil, false
	}
	t, ok := sc.expireTime[key]
	if ok && time.Now().After(t) {
		return nil, false
	}
	return val, true
}

// GetOrSetRefresh  获取值，如果key已经过期的话，那么设置刷新函数，和过期时间，并且重新获取值返回
// refreshDuration是value被cache自动刷新的间隔
// expireDuration指的是这个缓存key过期的时间, 正常情况下expireDuration应该远远大于refreshDuration, 0表示永不过期
// 注意： refresh函数的刷新操作并不会修改key本身的expireDuration时间
// TODO 可能需要增加一个方法来更新refresh 时间，因为交易时和非交易时的区别
func (sc *SimpleCache) GetOrSetRefresh(key string, loadFn func() (interface{}, error), expireDuration, refreshDuration time.Duration) (interface{}, error) {
	ret, ok := sc.get(key)
	if ok {
		return ret, nil
	}

	v, err, _ := sc.sfOnce.Do(key, func() (i interface{}, e error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic %+v", r)
				e = errors.New("GetOrSetRefresh loadFn panic")
			}
		}()
		ret, err := loadFn()
		if err != nil {
			return nil, err
		}
		expireTime := time.Now().Add(expireDuration)
		if expireDuration == 0 {
			expireTime = time.Now().AddDate(30, 0, 0) //过期时间设置为30年后，也就是永不过期
		}
		sc.set(key, ret, expireTime)
		// 下面设置refresh 函数
		sc.setRefreshFunc(key, loadFn, refreshDuration)
		return ret, err
	})
	return v, err
}

// setRefreshFunc 主动设置缓存key的后台刷新函数，这会使得之前的刷新函数退出
func (sc *SimpleCache) setRefreshFunc(key string, loadFn func() (interface{}, error), refreshDuration time.Duration) {
	sc.lock.Lock()
	defer sc.lock.Unlock()
	// 下面先关闭之前的chan， 本来想复用这个goroutine，但是考虑到要更新刷新频率, 所以还是需要重新建立协程
	oldCh, ok := sc.refreshChan[key]
	if ok {
		close(oldCh)
	}
	ch := make(chan interface{})
	sc.refreshChan[key] = ch
	log.Infof("update refresh func for cache key %s", key)

	go func() {
		log.Infof("enter refresh loop for cache key %s", key)
		tick := time.NewTicker(refreshDuration).C
		for {
			select {
			case <-tick:
				log.Infof("start to refresh key %s", key)
				data, err := loadFn()
				if err != nil {
					log.Errorf("failed to update cache key %s, err %+v", key, err)
					continue
				}
				// 这里没有用set函数就是为了不更新expire时间
				sc.lock.Lock()
				sc.data[key] = data
				sc.lock.Unlock()
				log.Infof("successfully refresh key %s", key)
			case <-ch:
				log.Infof("exit refresh loop for cache key %s", key)
				return
			}
		}
	}()
}

// GetAndLoad 最常用的功能，先获取，没有的话会更新缓存
func (sc *SimpleCache) GetAndLoad(key string, loadFn func() (interface{}, error), expireDuration time.Duration) (interface{}, error) {
	now := time.Now()
	ret, ok := sc.get(key)
	if ok {
		return ret, nil
	}
	v, err, _ := sc.sfOnce.Do(key, func() (i interface{}, e error) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic %+v", r)
				e = errors.New("GetAndLoad loadFn panic")
			}
		}()
		ret, err := loadFn()
		if err != nil {
			return nil, err
		}
		sc.set(key, ret, now.Add(expireDuration))
		return ret, nil
	})
	return v, err
}
