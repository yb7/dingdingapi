package svc

import (
  "crypto/sha1"
  "math/rand"
  "time"
  "fmt"
  "strconv"
  "encoding/json"
  "golang.org/x/net/context"
  "github.com/yb7/dingdingapi/pb"
  "github.com/yb7/dingdingapi/config"
  "github.com/yb7/dingdingapi/util"
)

// dingindg jssdk签名
type JssdkResult struct {
  Errcode      int    `json:"errcode"`
  Errmsg       string `json:"errmsg"`
  Access_token string `json:"access_token"`
  Expires_in   int    `json:"expires_in"`
  Ticket       string `json:"ticket"`
}

var URL_TOKEN = "https://oapi.dingtalk.com/gettoken"
var URL_TICKET = "https://oapi.dingtalk.com/get_jsapi_ticket"

func (s *DingDingService) GetJssdkSign(context context.Context, req *pb.GetJssdkSignRequest) (*pb.GetJssdkSignResponse, error) {
  return nil, nil
}

func GetJssdkSign(urls string) map[string]string {
  var log = util.AppLog.With("func", "GetJssdkSign")
  //var appId = ding_appid
  //var secret = ding_appsecret
  var tokenMap = JssdkResult{}
  var jsonTicket = JssdkResult{}
  var noncestr = getRandomString(20, 3)
  var timestamp = strconv.FormatInt(time.Now().Unix(), 10)
  var sign = map[string]string{}

  // 获取jsapiTicket从redis中7200秒，如果没有再请求
  var wxJsapiTicketKey = "dd_jsapi_ticket_7200"
  var wxTicket = ""
  if redisCache.Exists(wxJsapiTicketKey).Val() > 0 {
    wxTicket, _ = redisCache.Get(wxJsapiTicketKey).Result()
  } else {
    // 通过appid+secret获取accesstoken  https://oapi.dingtalk.com/gettoken?corpid=id&corpsecret=secrect
    jsapiToken, err := get(URL_TOKEN + "?corpid=" + config.CORP_ID + "&corpsecret=" + config.CORP_SECRET)
    if err != nil {
      log.Errorf("gettoken err = %v", err)
      return sign
    }
    err = json.Unmarshal(jsapiToken, &tokenMap)
    if err != nil {
      log.Errorf("Unmarshal jsapiToken", err)
      return sign
    }
    if tokenMap.Access_token == "" {
      log.Errorf("jsapiToken httpdo error")
      return sign
    }

    // 通过access_token获取ticket https://oapi.dingtalk.com/get_jsapi_ticket?access_token=ACCESS_TOKE
    jsapiTicket, err := get(URL_TICKET + "?access_token=" + tokenMap.Access_token)
    if err != nil {
      log.Errorf("get_jsapi_ticket err = %v", err)
      return sign
    }

    err = json.Unmarshal(jsapiTicket, &jsonTicket)
    if err != nil {
      log.Errorf("Unmarshal jsapiTicket", err)
      return sign
    }
    log.Debugf("unmarshal jsapiTicket = %v", string(jsapiTicket))

    // 如果成功获取ticket保存到redis 7200秒，否则返回错误
    if jsonTicket.Ticket != "" {
      err = redisCache.Set(wxJsapiTicketKey, jsonTicket.Ticket, time.Second*7200).Err()
      if err != nil {
        log.Errorf("redisCache.Set(wxJsapiTicketKey, jsonTicket.Ticket, time.Second * 7200)", err)
        return sign
      }
      wxTicket = jsonTicket.Ticket
    } else {
      log.Errorf("jsapiTicket httpdo error")
      return sign
    }
  }

  // 通过获取的ticket计算签名"jsapi_ticket=" + jsTicket +"&noncestr=" + nonce +"&timestamp=" + timeStamp + "&url=" + url;
  var sginstr = "jsapi_ticket=" + wxTicket + "&noncestr=" + noncestr + "&timestamp=" + timestamp + "&url=" + urls
  log.Infof("sginstr ", sginstr)
  signature := Js_sha1(sginstr)
  sign["agentId"] = config.AGENT_ID
  sign["corpId"] = config.CORP_ID
  sign["nonceStr"] = noncestr
  sign["timestamp"] = timestamp
  sign["signature"] = signature
  sign["url"] = urls
  log.Infof("ddTicket=", wxTicket)
  log.Infof("sign=", sign)
  return sign
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
