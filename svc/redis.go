package svc


import (
  "github.com/go-redis/redis"
  "github.com/yb7/dingdingapi/config"
  "github.com/yb7/dingdingapi/util"
)

var redisLog = util.AppLog.With("file", "service/redis.go")

var redisCache *redis.Client

func OpenRedis() *redis.Client {
  log := redisLog.With("func", "OpenRedis")
  redisCache = redis.NewClient(&redis.Options{Addr: config.REDIS_ADDR, Password: config.REDIS_PWD, DB: 0})
  sc := redisCache.Ping()
  log.Infof("Connectiong to %s: Ping => %s", redisCache.String(), sc.Val())
  if sc.Err() != nil {
    log.Errorf("OpenRedis conection set up failed, %s", sc.Err())
    panic(sc.Err())
  }
  log.Infof("Redis conection set up successfully")
  return redisCache
}

func CloseRedis() {
  redisCache.Close()
}
