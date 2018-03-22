package svc

import (
  "github.com/yb7/dingdingapi/config"
  "time"
  "github.com/yb7/dingdingapi/util"
  "fmt"
  "encoding/json"
)

var accessTokenLog = util.AppLog.With("file", "svc/access_token.go")

var appAccessToken string
var corpAccessToken string

func GetDingAccessTokenEveryHour() {
  log := accessTokenLog.With("func", "getDingAccessTokenEveryHour")

  var getCorpAccessToken = func() {
    if len(config.CORP_ID) > 0 && len(config.CORP_SECRET) > 0 {
      dtrMessage, err := getDingAccessToken(false)
      if err != nil {
        log.Errorf("getCorpAccessToken err %v", err)
        return
      }
      corpAccessToken = dtrMessage.AccessToken
      log.Infof("corp access token is %s", corpAccessToken)
    }
  }
  var getAppAccessToken = func() {
    dtr, err := getDingAccessToken(true)
    if err != nil {
      log.Warnf("ticker getAppAccessToken failed:%v", err)
      return
    }
    appAccessToken = dtr.AccessToken
    log.Infof("app access token is %s", appAccessToken)
  }

  getAppAccessToken()
  getCorpAccessToken()

  timer := time.NewTicker(1 * time.Hour)
  go func() {
    for {
      select {
      case <-timer.C:
        go func() {
          getAppAccessToken()
          getCorpAccessToken()
        }()
      }
    }
  }()
}


/**
 * 使用appid及appSecret访问如下接口，获取accesstoken，此处获取的token有效期为2小时，
 * 有效期内重复获取，返回相同值，并自动续期，如果在有效期外获取会获得新的token值，建议定时获取本token，不需要用户登录时再获取。
 */
func getDingAccessToken(isLoginToken bool) (resp DingTokenResp, err error) {
  log := dingMsgLog.With("func", "getDingAccessToken")
  var requestUrl string
  if isLoginToken == true {
    requestUrl = fmt.Sprintf(config.DING_HOST+"sns/gettoken?appid=%s&appsecret=%s", config.APP_ID, config.APP_SECRET)
  } else {
    requestUrl = fmt.Sprintf(config.DING_HOST+"gettoken?corpid=%s&corpsecret=%s", config.CORP_ID, config.CORP_SECRET)
  }

  result, err := get(requestUrl)
  if err != nil {
    err = fmt.Errorf("get err %v", err)
    log.Error(err)
    return
  }

  err = json.Unmarshal(result, &resp)
  if err != nil {
    err = fmt.Errorf("Unmarshal err %v", err)
    log.Error(err)
    return
  }
  if resp.ErrCode != 0 && resp.ErrMsg != "ok" {
    err = fmt.Errorf("ding err code: %v; ding err msg: %v", resp.ErrCode, resp.ErrMsg)
    log.Error(err)
    return
  }

  return
}
