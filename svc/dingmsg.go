package svc

import (
  "encoding/json"
  "github.com/yb7/dingdingapi/pbdingding"
  "time"
  "net/url"
  "strings"
  "golang.org/x/net/context"
  "github.com/yb7/dingdingapi/config"
  "github.com/yb7/dingdingapi/util"
)


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


type DingRespErr struct {
  ErrCode int    `json:"errcode"`
  ErrMsg  string `json:"errmsg"`
}
type DingTokenResp struct {
  DingRespErr
  AccessToken string `json:"access_token"`
}


func NewMessageParam(dingMessageMethod string) *MessageParam {
  mp := new(MessageParam)
  mp.Method = dingMessageMethod
  mp.Session = corpAccessToken
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

type DingSendMsgResp struct {
  Result *struct {
    ErrCode int32  `json:"ding_open_errcode"`
    ErrMsg  string `json:"error_msg"`
    Success bool   `json:"success"`
    TaskID  int64  `json:"task_id"`
  } `json:"result"`
  RequestId string `json:"request_id"`
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
  //log.Debugf(param)
  log.Debugf("content of sendMessage api >>> \nmethod: %s, session: %s, timestamp: %s, format: %s, v: %s\nmsgtype: %s, userid_list: %s, agent_id: %s\nmsgcontent: %s",
    mp.Method, mp.Session, mp.TimesTamp, mp.Format, mp.V, mp.MsgType, mp.UserIDList, mp.AgentID, mp.MsgContent)

  result, err := post(config.DING_MESSAGE_URL, []byte(param), false)
  if err != nil {
    return nil, err
  }
  var resp = &pbdingding.SendDingMessageResponse{}
  var m = make(map[string]DingSendMsgResp)

  log.Debugf("send message response: %s", string(result))

  if err := json.Unmarshal(result, &m); err != nil {
    log.Error(err)
  } else {
    for _, v := range m {
      if v.Result != nil {
        resp.DingOpenErrorCode = v.Result.ErrCode
        resp.ErrorMsg = v.Result.ErrMsg
        resp.Success = v.Result.Success
        resp.TaskID = v.Result.TaskID //v.requestId
      }
      resp.RequestID = v.RequestId
    }
  }

  return resp, nil
}
