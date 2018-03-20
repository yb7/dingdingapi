package svc

import (
  "crypto/sha1"
  "math/rand"
  "time"
  "fmt"
  "strconv"
  "encoding/json"
  "golang.org/x/net/context"
  "github.com/yb7/dingdingapi/pbdingding"
  "github.com/yb7/dingdingapi/config"
  "github.com/yb7/dingdingapi/util"
  "errors"
)

// dingindg jssdk签名
type JssdkResult struct {
  Errcode      int    `json:"errcode"`
  Errmsg       string `json:"errmsg"`
  Access_token string `json:"access_token"`
  Expires_in   int    `json:"expires_in"`
  Ticket       string `json:"ticket"`
}

func ddJsapiTicketKey() string {
  return fmt.Sprintf("%sdd_jsapi_ticket", config.REDIS_KEY_PREFIX)
}
var URL_TOKEN = "https://oapi.dingtalk.com/gettoken"
var URL_TICKET = "https://oapi.dingtalk.com/get_jsapi_ticket"
var jssdkSignLog = util.AppLog.With("file", "svc/jssdksign.go")

/**
 * https://open-doc.dingtalk.com/docs/doc.htm?spm=a219a.7629140.0.0.5kvQdT&treeId=385&articleId=104966&docType=1
 */
func GetDingJSTokenEveryHour() {
  log := dingMsgLog.With("func", "getDingJSTokenEveryHour")
  if len(dingMessageAccessToken) == 0 {
    log.Warnf("DingMessageAccessToken is blank, stop to get js token")
    return
  }
  _, err := getJSTicket()
  if err != nil {
    log.Errorf("getJSTicket err = %v", err)
    panic(err)
    return
  }

  timer := time.NewTicker(1 * time.Hour)
  go func() {
    for {
      select {
      case <-timer.C:
        go func() {
          _, err := getJSTicket()
          if err != nil {
            log.Errorf("getJSTicket err = %v", err)
            panic(err)
            return
          }
        }()
      }
    }
  }()
}

func getJSTicket() (string, error) {
  var log = jssdkSignLog.With("func", "getJSTicket")
  var jsonTicket = JssdkResult{}
  var ticket string
  jsapiTicket, err := get(URL_TICKET + "?access_token=" + dingMessageAccessToken)
  if err != nil {
    log.Errorf("get = %v", err)
    return ticket, err
  }

  err = json.Unmarshal(jsapiTicket, &jsonTicket)
  if err != nil {
    log.Errorf("json.Unmarshal err = %v", err)
    return ticket, err
  }
  log.Debugf("unmarshal jsapiTicket = %v", string(jsapiTicket))

  err = TokenCache.Set(ddJsapiTicketKey(), jsonTicket.Ticket, time.Second*3600)
  if err != nil {
    log.Errorf("redisCache.Set err = %v", err)
    return ticket, err
  }
  ticket = jsonTicket.Ticket
  return ticket, nil
}

func (s *DingDingService) GetJssdkSign(context context.Context, req *pbdingding.GetJssdkSignRequest) (*pbdingding.GetJssdkSignResponse, error) {
  if len(config.AGENT_ID) == 0 || len(config.CORP_ID) == 0 {
    return nil, errors.New("missing config key AGENT_ID/CORP_ID")
  }
  var log = jssdkSignLog.With("func", "GetJssdkSign")
  var noncestr = getRandomString(20, 3)
  var timestamp = strconv.FormatInt(time.Now().Unix(), 10)
  var sign = &pbdingding.GetJssdkSignResponse{}
  var ddTicket string
  ddTicket, _ = TokenCache.Get(ddJsapiTicketKey())

  var sginstr = "jsapi_ticket=" + ddTicket + "&noncestr=" + noncestr + "&timestamp=" + timestamp + "&url=" + req.Url
  log.Infof("sginstr ", sginstr)
  signature := Js_sha1(sginstr)
  sign.AgentId = config.AGENT_ID
  sign.CorpId = config.CORP_ID
  sign.NonceStr = noncestr
  sign.Timestamp = timestamp
  sign.Signature = signature
  sign.Url = req.Url
  log.Infof("ddTicket=", ddTicket)
  log.Infof("sign=", sign)
  return sign, nil
}

/**
 * 生成随机字符串
 * @param  num int
 * @param  kind
    KC_RAND_KIND_NUM   = 0 // 纯数字
    KC_RAND_KIND_LOWER = 1 // 小写字母
    KC_RAND_KIND_UPPER = 2 // 大写字母
    KC_RAND_KIND_ALL   = 3 // 数字、大小写字母
 * @return str string
 */
func getRandomString(size int, kind int) string {
  ikind, kinds, result := kind, [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)
  is_all := kind > 2 || kind < 0
  rand.Seed(time.Now().UnixNano())
  for i := 0; i < size; i++ {
    if is_all { // random ikind
      ikind = rand.Intn(3)
    }
    scope, base := kinds[ikind][0], kinds[ikind][1]
    result[i] = uint8(base + rand.Intn(scope))
  }
  return string(result)
}

/*
* 生成sha1
*anthor:vance
*/
func Js_sha1(data string) string {
  h := sha1.New()
  h.Write([]byte(data))
  sha1str1 := h.Sum(nil)
  sha1str2 := fmt.Sprintf("%x", sha1str1)
  return sha1str2

  /*AWSSecretKeyId := "ooxxooxx"
  sha256 := sha256.New
  hash := hmac.New(sha256, []byte(AWSSecretKeyId))
  hash.Write([]byte(data))
  sha := base64.StdEncoding.EncodeToString(hash.Sum(nil))
  sha= url.QueryEscape(sha)
  return sha*/
}
