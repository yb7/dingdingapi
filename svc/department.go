package svc

import (
  "golang.org/x/net/context"
  "github.com/yb7/dingdingapi/pbdingding"
  "fmt"
  "encoding/json"
  "errors"
)

type DingDingDepartmentsResp struct {
  DingRespErr
  Department []*pbdingding.Departments_Department
}
type UsersInDepartmentResp struct {
  DingRespErr
  HasMore bool
  UserList []*pbdingding.UsersInDepartment_User
}
func (s *DingDingService) GetDepartments(ctx context.Context, req *pbdingding.GetDepartmentsRequest) (*pbdingding.Departments, error) {
  bytes, err := get(fmt.Sprintf("https://oapi.dingtalk.com/department/list?access_token=%s", corpAccessToken))
  if err != nil {
    return nil, err
  }
  var resp = DingDingDepartmentsResp{}
  err = json.Unmarshal(bytes, &resp)
  if err != nil {
    return nil, err
  }
  if resp.ErrCode == 0 {
    return &pbdingding.Departments{
      Departments: resp.Department,
    }, nil
  }
  return nil, errors.New(resp.ErrMsg)
}


func (*DingDingService) GetUsersInDepartment(ctx context.Context, req *pbdingding.GetUsersInDepartmentRequest) (*pbdingding.UsersInDepartment, error) {
  bytes, err := get(fmt.Sprintf("https://oapi.dingtalk.com/user/simplelist?access_token=%s&department_id=%d", corpAccessToken, req.DepartmentID))
  if err != nil {
    return nil, err
  }
  var resp = UsersInDepartmentResp{}
  err = json.Unmarshal(bytes, &resp)
  if err != nil {
    return nil, err
  }
  if resp.ErrCode == 0 {
    return &pbdingding.UsersInDepartment{
      Users: resp.UserList,
    }, nil
  }
  return nil, errors.New(resp.ErrMsg)
}
