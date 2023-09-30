package db

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"time"

	"stock/common/env"
	"stock/common/log"

	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

// InitRedisClient 初始化redis client
func InitRedisClient() {

	// 解析redis配置
	redisConfStr, ok := env.GlobalEnv().Get("REDIS")
	if !ok {
		panic("no REDIS config")
	}
	// redisConfStr: redis://<user>:<password>@<host>:<port>/<db_number>
	conf, err := redis.ParseURL(redisConfStr)
	if err != nil {
		panic(fmt.Errorf("parse REDIS config failed: %+v", err))
	}
	rc = redis.NewClient(conf)

	err = checkRedisConnect(context.Background())
	if err != nil {
		panic(fmt.Errorf("check redis connect failed: %+v", err))
	}
	log.Infof("connect to redis success")
}

var rc *redis.Client

// RedisClient 获取redis client
func RedisClient() *redis.Client {
	return rc
}

// GetOrLoad 从redis cache 中加载或者使用loader函数加载
func GetOrLoad(ctx context.Context, key string, expire time.Duration, holder interface{}, loader func() error) error {
	rv := reflect.ValueOf(holder)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(holder)}
	}
	// get from redis
	buf, err := RedisClient().Get(ctx, key).Bytes()
	if err == nil {
		// log.Debugf("load %s from redis", key)
		// get result from redis
		return json.Unmarshal(buf, holder)
	}
	if err != redis.Nil {
		// 直接失败
		return err
	}
	// err == redis.Nil
	// load result from db
	_, err, _ = sfOnce.Do(key, func() (interface{}, error) {
		return nil, loader()
	})

	if err != nil {
		log.Errorf("get %s from db, err: %v", key, err)
		return err
	}
	// set result to cache
	buf, err = json.Marshal(holder)
	if err != nil {
		log.Errorf("marshal json err: %v", err)
		return err
	}
	return RedisClient().Set(ctx, key, buf, expire).Err()
}

var sfOnce singleflight.Group

func checkRedisConnect(ctx context.Context) error {
	if rc == nil {
		return fmt.Errorf("client has not created")
	}

	return rc.Set(ctx, "redis_check_connect", 1, time.Minute).Err()
}

// Set 封装redis client的Set方法
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return RedisClient().Set(ctx, key, value, expiration)
}

// Get 封装redis client的Get方法
func Get(ctx context.Context, key string) *redis.StringCmd {
	return RedisClient().Get(ctx, key)
}

// Delete 封装redis的Del
func Delete(ctx context.Context, key string) *redis.IntCmd {
	return RedisClient().Del(ctx, key)
}

// MGet 封装redis client的MGet方法
func MGet(ctx context.Context, key ...string) *redis.SliceCmd {
	return RedisClient().MGet(ctx, key...)
}

// MSet 封装redis client的MSet方法
func MSet(ctx context.Context, value ...interface{}) *redis.StatusCmd {
	return RedisClient().MSet(ctx, value)
}

// MGetFromRedis 从redis批量取数据，集群模式下必须使用hashtag
func MGetFromRedis(ctx context.Context, holderMap map[string]interface{}, compressed bool) error {
	keys := make([]string, 0, len(holderMap))
	for k := range holderMap {
		keys = append(keys, k)
	}

	ret, err := RedisClient().MGet(ctx, keys...).Result()
	if err != nil {
		return err
	}
	for i, key := range keys {
		if ret[i] == nil {
			holderMap[key] = nil
			continue
		}
		v, ok := ret[i].(string)
		if !ok {
			continue
		}
		holder, ok := holderMap[key]
		if !ok {
			continue
		}
		err := unmarshal([]byte(v), holder, compressed)
		if err != nil {
			return err
		}
	}
	return nil
}

func unmarshal(buf []byte, holder interface{}, compressed bool) error {
	rv := reflect.ValueOf(holder)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("invalid type %v", reflect.TypeOf(holder))
	}
	if !compressed {
		return json.Unmarshal(buf, holder)
	}
	gr, err := gzip.NewReader(bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer gr.Close()
	out, err := ioutil.ReadAll(gr)
	if err != nil {
		return err
	}
	return json.Unmarshal(out, holder)
}

// MSaveToRedis 批量存数据到redis，集群模式下必须使用hashtag
func MSaveToRedis(data map[string]interface{}, expire time.Duration, shouldCompress bool) error {
	ctx := context.Background()
	if len(data) == 0 {
		return nil
	}
	m := make(map[string]interface{}, len(data))
	for k, v := range data {
		buf, err := marshal(v, shouldCompress)
		if err != nil {
			return err
		}
		m[k] = buf
	}

	_, err := RedisClient().MSet(ctx, m).Result()
	if err != nil {
		return err
	}

	pipeline := RedisClient().Pipeline()
	for key := range data {
		expire += time.Duration(rand.Intn(100)) * time.Millisecond
		_ = pipeline.Expire(ctx, key, expire)
	}

	_, err = pipeline.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func marshal(data interface{}, shouldCompress bool) ([]byte, error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if !shouldCompress {
		return buf, nil
	}

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err = gz.Write(buf)
	if err != nil {
		gz.Close()
		return nil, err
	}
	// 必须先关gzWriter才能开始读，不然有一部分数据不会被flush到buffer
	gz.Close()
	return b.Bytes(), nil
}
