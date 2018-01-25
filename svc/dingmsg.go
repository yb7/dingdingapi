package svc

import (
  "encoding/json"
  "github.com/yb7/dingdingapi/pbdingding"
  "time"
  "fmt"
  "net/url"
  "strings"
  "golang.org/x/net/context"
  "github.com/yb7/dingdingapi/config"
  "github.com/yb7/dingdingapi/util"
)

//type MessageContent struct {
//  Body struct {
//    Author    string `json:"author"`
//    Content   string `json:"content"`
//    FileCount string `json:"file_count"`
//    Form []struct {
//      Key   string `json:"key"`
//      Value string `json:"value"`
//    } `json:"form"`
//    Image string `json:"image"`
//    Rich struct {
//      Num  string `json:"num"`
//      Unit string `json:"unit"`
//    } `json:"rich"`
//    Title string `json:"title"`
//  } `json:"body"`
//  Head struct {
//    Bgcolor string `json:"bgcolor"`
//    Text    string `json:"text"`
//  } `json:"head"`
//  MessageURL string `json:"message_url"`
//}

var dingMsgLog = util.AppLog.With("file", "svc/dingmsg.go")
func toBytes(p *pbdingding.SendDingMessageRequest_Content) ([]byte, error) {
  //dingLog.Debugf("MessageContent = %+v", *p)
  content, err := json.Marshal(p)
  if err != nil {
    return nil, err
  }
  return content, nil
}

type MessageParam struct {
  Method     string
  Session    string
  TimesTamp  string
  Format     string
  V          string
  MsgType    string
  AgentID    string
  UserIDList string
  MsgContent string
}
var dingAccessToken string
var dingMessageAccessToken string

func init() {
  getDingAccessTokenEveryHour()
}
func getDingAccessTokenEveryHour() {
  log := dingMsgLog.With("func", "getDingAccessTokenEveryHour")
  dtr, err := getDingAccessToken(true)
  if err != nil {
    panic(err)
    return
  }
  //log.Infof("access_token = %v", dtr.AccessToken)
  dingAccessToken = dtr.AccessToken

  dtrMessage, err := getDingAccessToken(false)
  if err != nil {
    panic(err)
  }
  //log.Infof("message_access_token = %v", dtrMessage.AccessToken)
  dingMessageAccessToken = dtrMessage.AccessToken

  timer := time.NewTicker(1 * time.Hour)
  go func() {
    for {
      select {
      case <-timer.C:
        go func() {
          dtr, err := getDingAccessToken(true)
          if err != nil {
            log.Warnf("ticker getDingAccessToken failed:%v", err)
            return
          }
          dingAccessToken = dtr.AccessToken

          dtrMessage, err := getDingAccessToken(false)
          if err != nil {
            log.Errorf("getDingAccessToken err %v", err)
            return
          }
          dingMessageAccessToken = dtrMessage.AccessToken
        }()
      }
    }
  }()
}
type DingRespErr struct {
  ErrCode int    `json:"errcode"`
  ErrMsg  string `json:"errmsg"`
}
type DingTokenResp struct {
  DingRespErr
  AccessToken string `json:"access_token"`
}
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

func NewMessageParam(dingMessageMethod string) *MessageParam {
  mp := new(MessageParam)
  mp.Method = dingMessageMethod
  mp.Session = dingMessageAccessToken
  mp.TimesTamp = time.Now().Format("2006-01-02 15:04:05")
  mp.Format = "json"
  mp.V = "2.0"
  mp.MsgType = "oa"
  mp.AgentID = config.AGENT_ID
  return mp
}
func (p *MessageParam) FormEncoded() string {
  v := url.Values{}
  v.Set("method", p.Method)
  v.Add("session", p.Session)
  v.Add("timestamp", p.TimesTamp)
  v.Add("format", p.Format)
  v.Add("v", p.V)
  v.Add("msgtype", p.MsgType)
  v.Add("userid_list", p.UserIDList)
  v.Add("msgcontent", p.MsgContent)
  v.Add("agent_id", p.AgentID)
  result := v.Encode()
  return result
}

type DingDingService struct {

}

type dingSendMsgResp struct {
  result *struct {
    errCode int32 `json:"ding_open_errcode"`
    errMsg  string `json:"error_msg"`
    success  bool `json:"success"`
  }
  requestId string `json:"request_id"`
}
// https://open-doc.dingtalk.com/docs/doc.htm?spm=a219a.7629140.0.0.ccmVn3&treeId=385&articleId=28915&docType=2
// dingtalk.corp.message.corpconversation.asyncsend (企业会话消息异步发送)
func (s *DingDingService) SendMessage(ctx context.Context, req *pbdingding.SendDingMessageRequest) (*pbdingding.SendDingMessageResponse, error) {
  log := util.AppLog.With("func", "dingMessage")

  messageContent, err := toBytes(req.Content)
  if err != nil {
    return nil, err
  }

  mp := NewMessageParam(req.Method)
  mp.UserIDList = strings.Join(req.Recipients, ",")

  mp.MsgContent = string(messageContent)
  log.Debugf(mp.MsgContent)
  param := mp.FormEncoded()
  log.Debugf(param)

  result, err := post(config.DING_MESSAGE_URL, []byte(param), false)
  if err != nil {
    return nil, err
  }
  var resp = &pbdingding.SendDingMessageResponse{}
  var m = make(map[string]dingSendMsgResp)

  log.Debugf("send message response: %s", string(result))

  if err := json.Unmarshal(result, &m); err != nil {
    log.Error(err)
  } else {
    for _, v := range m {
      resp.DingOpenErrorCode = v.result.errCode
      resp.ErrorMsg = v.result.errMsg
      resp.Success = v.result.success
      resp.TaskID = 0 //v.requestId
    }
  }
  log.Debugf("dingMessage result = %v", string(result))

  return resp, nil
}
