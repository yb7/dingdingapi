package svc


import (
  "github.com/go-redis/redis"
  "github.com/yb7/dingdingapi/config"
  "github.com/yb7/dingdingapi/util"
  "time"
)

var redisLog = util.AppLog.With("file", "service/redis.go")

var TokenCache interface {
  Set(key string, value string, expiration time.Duration) error
  Get(key string) (string, bool)
}
type InMemoryTokenCache struct {
  cachedValue map[string]CacheValue
}
type CacheValue struct {
  Value string
  ExpiresAt time.Time
}

func (c *InMemoryTokenCache) Set(key string, value string, expiration time.Duration) error {
  if c.cachedValue == nil {
    c.cachedValue = make(map[string]CacheValue)
  }
  c.cachedValue[key] = CacheValue{
    Value: value,
    ExpiresAt: time.Now().Add(expiration),
  }
  return nil
}

func (c *InMemoryTokenCache) Get(key string) (string, bool) {
  if c.cachedValue == nil {
    return "", false
  }
  v, ok := c.cachedValue[key]
  if !ok {
    return "", false
  }
  if v.ExpiresAt.Before(time.Now()) {
    delete(c.cachedValue, key)
    return "", false
  }
  return v.Value, true
}

type RedisTokenCache struct {
  redisCache *redis.Client
}
func (c *RedisTokenCache) Set(key string, value string, expiration time.Duration) error {
  _, err := c.redisCache.Set(key, value, expiration).Result()
  return err
}
func (c *RedisTokenCache) Get(key string) (string, bool) {
  str, err := c.redisCache.Get(key).Result()
  return str, err == nil
}

func OpenTokenCache() {
  log := redisLog.With("func", "OpenTokenCache")
  if len(config.REDIS_ADDR) > 0 {
    redisCache := redis.NewClient(&redis.Options{Addr: config.REDIS_ADDR, Password: config.REDIS_PWD, DB: 0})
    sc := redisCache.Ping()
    log.Infof("Connectiong to %s: Ping => %s", redisCache.String(), sc.Val())
    if sc.Err() != nil {
      log.Errorf("OpenRedis conection set up failed, %s", sc.Err())
      panic(sc.Err())
    }
    log.Infof("Redis conection set up successfully")
    TokenCache = &RedisTokenCache{
      redisCache: redisCache,
    }
  } else {
    TokenCache = &InMemoryTokenCache{}
  }

}

func CloseTokenCache() {
  log := redisLog.With("func", "OpenTokenCache")
  switch v := TokenCache.(type) {
  case *RedisTokenCache:
    log.Infof("close redis token cache")
    v.redisCache.Close()
  case *InMemoryTokenCache:
    log.Infof("close in memory token cache")
  default:
    log.Infof("unknown cache %T", v)
  }
}

//func OpenRedis() *redis.Client {
//  log := redisLog.With("func", "OpenRedis")
//  redisCache := redis.NewClient(&redis.Options{Addr: config.REDIS_ADDR, Password: config.REDIS_PWD, DB: 0})
//  sc := redisCache.Ping()
//  log.Infof("Connectiong to %s: Ping => %s", redisCache.String(), sc.Val())
//  if sc.Err() != nil {
//    log.Errorf("OpenRedis conection set up failed, %s", sc.Err())
//    panic(sc.Err())
//  }
//  log.Infof("Redis conection set up successfully")
//  return redisCache
//}
//
//func CloseRedis() {
//  redisCache.Close()
//}
