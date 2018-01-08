package util

import (
  "github.com/yb7/alilog"
  "github.com/yb7/dingdingapi/config"
)

var AppLog = alilog.New(config.LOG_PROJECT, config.LOG_STORE)
