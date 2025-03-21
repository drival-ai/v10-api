package main

import (
	"context"
	"sync"

	"github.com/drival-ai/v10-api/global"
	"github.com/drival-ai/v10-api/internal"
	iampb "github.com/drival-ai/v10-go/iam/v1"

	b "github.com/drival-ai/v10-api/services/base"
	iam "github.com/drival-ai/v10-api/services/iam"
	base "github.com/drival-ai/v10-go/base/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	cctx = func(p context.Context) context.Context {
		return context.WithValue(p, struct{}{}, nil)
	}
)

type service struct {
	ctx        context.Context
	clientOnce sync.Once
	UserInfo   internal.UserInfo
	Config     *global.Config

	iampb.UnimplementedIamServer
	base.UnimplementedV10Server
}

func (s *service) Register(ctx context.Context, req *iampb.RegisterRequest) (*iampb.RegisterResponse, error) {
	return iam.New().Register(ctx, req)
}

func (s *service) Login(ctx context.Context, req *iampb.LoginRequest) (*iampb.LoginResponse, error) {
	return iam.New().Login(ctx, req)
}

func (s *service) WhoAmI(ctx context.Context, req *iampb.WhoAmIRequest) (*iampb.WhoAmIResponse, error) {
	return iam.New().WhoAmI(ctx, req)
}

func (s *service) RegisterVehicle(ctx context.Context, req *base.RegisterVehicleRequest) (*emptypb.Empty, error) {
	return b.New().RegisterVehicle(ctx, req)
}
