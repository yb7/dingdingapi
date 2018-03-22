package svc

import (
  "time"
  "encoding/json"
  "github.com/yb7/dingdingapi/util"
  "fmt"
  "golang.org/x/net/context"
  "github.com/yb7/dingdingapi/pbdingding"
)

const (
  ding_host = "https://oapi.dingtalk.com/"
)

var dingLoginLog = util.AppLog.With("file", "svc/ding_login.go")

type (
  DingPersistentCodeResp struct {
    DingRespErr
    UnionID        string `json:"unionid"`
    OpenID         string `json:"openid"`
    PersistentCode string `json:"persistent_code"`
  }

  DingSNSTokenReq struct {
    OpenID         string `json:"openid"`
    PersistentCode string `json:"persistent_code"`
  }

  DingSNSTokenResp struct {
    DingRespErr
    SNSToken  string `json:"sns_token"`
    ExpiresIn int    `json:"expires_in"`
  }

  DingUserInfoResp struct {
    DingRespErr
    UserInfo DingUserInfo `json:"user_info"`
  }

  DingUserInfo struct {
    MaskedMobile string `json:"maskedMobile"`
    Nick         string `json:"nick"`
    UnionID      string `json:"unionid"`
    DingID       string `json:"dingId"`
    OpenID       string `json:"openid"`
  }

  DingPersistentCodeReq struct {
    TmpAuthCode string `json:"tmp_auth_code"`
  }
)

func (s *DingDingService) GetDingUserInfo(ctx context.Context, req *pbdingding.GetDingUserInfoRequest) (*pbdingding.GetDingUserInfoResponse, error) {
  log := dingLoginLog.With("func", "GetDingUserByQRCode")
  resp := &pbdingding.GetDingUserInfoResponse{}

  dingPersistentCodeResp, err := getDingPersistentCode(req.TmpAuthCode)
  if err != nil {
    log.Errorf("getDingPersistentCode err = %v", err)
    return resp, err
  }

  dingSNSTokenResp, err := getDingSNSToken(dingPersistentCodeResp)
  if err != nil {
    log.Errorf("getDingSNSToken err = %v", err)
    return resp, err
  }

  dingUserInfoResp, err := getDingUserInfo(dingSNSTokenResp)
  if err != nil {
    log.Errorf("getDingUserInfo err = %v", err)
    return resp, err
  }

  userInfo := dingUserInfoResp.UserInfo

  resp.MaskedMobile = userInfo.MaskedMobile
  resp.Nick = userInfo.Nick
  resp.Unionid = userInfo.UnionID
  resp.DingId = userInfo.DingID
  resp.Openid = userInfo.OpenID
  resp.SnsToken = dingSNSTokenResp.SNSToken

  nick := dingUserInfoResp.UserInfo.Nick

  //redis tokenCache
  TokenCache.Set(dingSNSTokenResp.SNSToken, nick, time.Hour*12)

  return resp, nil
}

func getDingPersistentCode(tmpAuthCode string) (resp DingPersistentCodeResp, err error) {
  requestUrl := fmt.Sprintf(ding_host+"/sns/get_persistent_code?access_token=%s", appAccessToken)
  dpcr := DingPersistentCodeReq{TmpAuthCode: tmpAuthCode}
  reqBody, err := json.Marshal(dpcr)
  if err != nil {
    err = fmt.Errorf("json.Marshal err: %v", err)
    return
  }

  result, err := post(requestUrl, reqBody, true)
  if err != nil {
    return
  }

  err = json.Unmarshal(result, &resp)
  if err != nil {
    err = fmt.Errorf("unmarshal err: %v", err)
    return
  }

  if resp.ErrCode != 0 && resp.ErrMsg != "ok" {
    err = fmt.Errorf("getDingPersistentCode err (ding err code: %v; ding err msg: %v)", resp.ErrCode, resp.ErrMsg)
    return
  }
  return
}

func getDingSNSToken(dpcr DingPersistentCodeResp) (resp DingSNSTokenResp, err error) {
  requestUrl := fmt.Sprintf(ding_host+"/sns/get_sns_token?access_token=%s", appAccessToken)
  dstr := DingSNSTokenReq{OpenID: dpcr.OpenID, PersistentCode: dpcr.PersistentCode}
  reqBody, err := json.Marshal(dstr)
  if err != nil {
    return
  }

  result, err := post(requestUrl, reqBody, true)
  if err != nil {
    err = fmt.Errorf("post failed: %v", err)
    return
  }

  err = json.Unmarshal(result, &resp)
  if err != nil {
    err = fmt.Errorf("unmarshal err %v", err)
    return
  }

  if resp.ErrCode != 0 && resp.ErrMsg != "ok" {
    err = fmt.Errorf("getDingSNSToken err (ding err code: %v; ding err msg: %v)", resp.ErrCode, resp.ErrMsg)
    return
  }
  return
}

func getDingUserInfo(dstr DingSNSTokenResp) (resp DingUserInfoResp, err error) {
  requestUrl := fmt.Sprintf(ding_host+"/sns/getuserinfo?sns_token=%s", dstr.SNSToken)
  result, err := get(requestUrl)
  if err != nil {
    err = fmt.Errorf("get err %v", err)
    return
  }

  err = json.Unmarshal(result, &resp)
  if err != nil {
    err = fmt.Errorf("unmarshal err %v", err)
    return
  }

  if resp.ErrCode != 0 && resp.ErrMsg != "ok" {
    err = fmt.Errorf("getDingUserInfo err (ding err code: %v; ding err msg: %v)", resp.ErrCode, resp.ErrMsg)
    return
  }

  return
}
