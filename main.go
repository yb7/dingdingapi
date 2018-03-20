package main

import (
  "github.com/yb7/dingdingapi/svc"
  "google.golang.org/grpc"
  "google.golang.org/grpc/reflection"
  "net"
  "github.com/yb7/dingdingapi/config"
  "github.com/yb7/dingdingapi/util"
  "github.com/yb7/dingdingapi/pbdingding"
)

var mainLog = util.AppLog.With("file", "main.go")
func main() {
  svc.OpenTokenCache()
  defer svc.CloseTokenCache()

  svc.GetDingAccessTokenEveryHour()

  svc.GetDingJSTokenEveryHour()

  grpcServer := grpc.NewServer()
  pbdingding.RegisterDingDingServer(grpcServer, &svc.DingDingService{})
  // Register reflection service on gRPC server.
  reflection.Register(grpcServer)

  lis, err := net.Listen("tcp", config.APP_PORT)
  if err != nil {
    mainLog.Errorf("failed to listen: %v", err)
  }

  mainLog.Infof("server started at port %s", config.APP_PORT)

  if err := grpcServer.Serve(lis); err != nil {
    mainLog.Errorf("failed to serve: %v", err)
  }
}
