package main

import (
	"context"
	"crypto/rsa"

	"github.com/drival-ai/v10-api/global"
	"github.com/drival-ai/v10-api/internal"
	iampb "github.com/drival-ai/v10-go/iam/v1"
	"github.com/golang/glog"

	b "github.com/drival-ai/v10-api/services/base"
	iam "github.com/drival-ai/v10-api/services/iam"
	base "github.com/drival-ai/v10-go/base/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	cctx = func(p context.Context) context.Context {
		return context.WithValue(p, struct{}{}, nil)
	}
)

type service struct {
	ctx        context.Context
	UserInfo   internal.UserInfo
	Config     *global.Config
	PrivateKey *rsa.PrivateKey

	iampb.UnimplementedIamServer
	base.UnimplementedV10Server
}

func (s *service) Register(ctx context.Context, req *iampb.RegisterRequest) (*iampb.RegisterResponse, error) {
	config := iam.Config{UserInfo: s.UserInfo, Config: s.Config, PrivateKey: s.PrivateKey}
	return iam.New(&config).Register(ctx, req)
}

func (s *service) Login(ctx context.Context, req *iampb.LoginRequest) (*iampb.LoginResponse, error) {
	config := iam.Config{UserInfo: s.UserInfo, Config: s.Config, PrivateKey: s.PrivateKey}
	return iam.New(&config).Login(ctx, req)
}

func (s *service) WhoAmI(ctx context.Context, req *iampb.WhoAmIRequest) (*iampb.WhoAmIResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	glog.Infof("md=%v", md)
	config := iam.Config{UserInfo: s.UserInfo, Config: s.Config, PrivateKey: s.PrivateKey}
	return iam.New(&config).WhoAmI(ctx, req)
}

func (s *service) RegisterVehicle(ctx context.Context, req *base.RegisterVehicleRequest) (*emptypb.Empty, error) {
	return b.New().RegisterVehicle(ctx, req)
}
