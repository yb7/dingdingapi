package config

import (
  "fmt"
  "os"
)

//type DingDingConfig struct {
//  AUTH_REDIRECT_URL string
//  APP_ID           string
//  AGENT_ID         string
//  APP_SECRET       string
//  CORP_ID          string
//  CORP_SECRET      string
//}

var (
  APP_PORT          string
  LOG_PROJECT       string
  LOG_STORE         string
  AUTH_REDIRECT_URL string
  APP_ID            string
  AGENT_ID          string
  APP_SECRET        string
  CORP_ID           string
  CORP_SECRET       string
  REDIS_ADDR        string
  REDIS_PWD         string
)

const DING_HOST = "https://oapi.dingtalk.com/"
const DING_MESSAGE_URL = "https://eco.taobao.com/router/rest"

//var config Config

type RedisConfig struct {
  Addr     string
  Password string
}


func init() {
  APP_PORT = mustString("APP_PORT")
  LOG_PROJECT = mustString("LOG_PROJECT")
  LOG_STORE = mustString("LOG_STORE")
  AUTH_REDIRECT_URL = mustString("AUTH_REDIRECT_URL")
  APP_ID = mustString("APP_ID")
  AGENT_ID = mustString("AGENT_ID")
  APP_SECRET = mustString("APP_SECRET")
  CORP_ID = mustString("CORP_ID")
  CORP_SECRET = mustString("CORP_SECRET")
  REDIS_ADDR = mustString("REDIS_ADDR")
  REDIS_PWD = mustString("REDIS_PWD")
}

//var configFilePath = os.Getenv("CONFIG_FILE")
//if len(configFilePath) == 0 {
//  panic("missing env key CONFIG_FILE")
//}
//data, err := ioutil.ReadFile(configFilePath)
//if err != nil {
//  panic(fmt.Errorf("error[%s] when read config file: %s", err.Error(), configFilePath))
//}
//err = json.Unmarshal(data, &config)
//if err != nil {
//  panic(fmt.Errorf("error[%s] when unmarshal config\n %s", err.Error(), string(data)))
//}
//AppLog = alilog.New(config.LOG_PROJECT, config.LOG_STORE)

//
func mustString(key string) string {
  var v = os.Getenv(key)
  if len(v) == 0 {
    panic(fmt.Sprintf("missing env key %s", key))
  }
  return v
}

//
//func readConfig() {
//  data, err := ioutil.ReadFile(mustString("CONFIG_FILE"))
//  if err != nil {
//    panic(fmt.Errorf("error[%s] when read sls config file: %s", err.Error(), mustString("CONFIG_FILE")))
//  }
//  var c config
//  err = json.Unmarshal(data, &c)
//  if err != nil {
//    panic(fmt.Errorf("error[%s] when unmarshal sls config\n %s", err.Error(), string(data)))
//  }
//  APP_PORT = c.APP_PORT
//  LOG_PROJECT = c.LOG_PROJECT
//  LOG_STORE = c.LOG_STORE
//  RpcServices = c.RpcServices
//  DingDing = c.DingDing
//  Redis = c.Redis
//}
