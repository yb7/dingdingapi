package svc

import (
  "io/ioutil"
  "bytes"
  "net/http"
  "errors"
  "github.com/yb7/dingdingapi/util"
)

func get(requestUrl string) ([]byte, error) {
  log := util.AppLog.With("func", "get")
  client := &http.Client{}
  req, err := http.NewRequest("GET", requestUrl, nil)
  if err != nil {
    log.Errorf("NewRequest err: %v\n", err)
    return nil, err
  }
  resp, err := client.Do(req)
  if err != nil {
    if resp != nil {
      resp.Body.Close()
    }
    log.Errorf("Do err: %v\n", err)
    return nil, err
  }
  if resp.StatusCode != 200 {
    return nil, err
  }
  defer resp.Body.Close()
  result, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Errorf("ioutil.ReadAll err: %v\n", err)
    return nil, err
  }
  return result, nil
}

func post(requestUrl string, data []byte, isJson bool) ([]byte, error) {
  log := util.AppLog.With("func", "post")
  client := &http.Client{}
  req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(data))
  req.Header.Add("Charset", "UTF-8")
  if isJson == true {
    req.Header.Add("Content-Type", "application/json")
  } else {
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  }

  resp, err := client.Do(req)

  if err != nil {
    if resp != nil {
      resp.Body.Close()
    }
    log.Warnf("Do err = %v", err)
    return nil, err
  }
  if resp == nil {
    return nil, nil
  }
  if resp.StatusCode != 200 {
    err = errors.New("resp status code != 200")
    log.Error(err)
    return nil, err
  }

  defer resp.Body.Close()
  r, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Warnf("ReadAll err = %v", err)
    return nil, err
  }
  return r, nil
}
