package main

import (
	"context"
	"crypto/rsa"

	"github.com/drival-ai/v10-api/global"
	"github.com/drival-ai/v10-api/internal"
	iampb "github.com/drival-ai/v10-go/iam/v1"

	base "github.com/drival-ai/v10-api/services/base"
	iam "github.com/drival-ai/v10-api/services/iam"
	basepb "github.com/drival-ai/v10-go/base/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	cctx = func(p context.Context) context.Context {
		return context.WithValue(p, struct{}{}, nil)
	}
)

type service struct {
	ctx        context.Context
	Config     *global.Config
	PrivateKey *rsa.PrivateKey

	iampb.UnimplementedIamServer
	basepb.UnimplementedV10Server
}

func (s *service) Register(ctx context.Context, req *iampb.RegisterRequest) (*iampb.RegisterResponse, error) {
	config := iam.Config{Config: s.Config, PrivateKey: s.PrivateKey}
	return iam.New(&config).Register(ctx, req)
}

func (s *service) Login(ctx context.Context, req *iampb.LoginRequest) (*iampb.LoginResponse, error) {
	config := iam.Config{Config: s.Config, PrivateKey: s.PrivateKey}
	return iam.New(&config).Login(ctx, req)
}

func (s *service) WhoAmI(ctx context.Context, req *iampb.WhoAmIRequest) (*iampb.WhoAmIResponse, error) {
	id := ctx.Value(internal.CtxKeyId)
	email := ctx.Value(internal.CtxKeyEmail)
	name := ctx.Value(internal.CtxKeyName)
	config := iam.Config{
		UserInfo: internal.UserInfo{
			Id:    id.(string),
			Email: email.(string),
			Name:  name.(string),
		},
		Config:     s.Config,
		PrivateKey: s.PrivateKey,
	}

	return iam.New(&config).WhoAmI(ctx, req)
}

func (s *service) RegisterVehicle(ctx context.Context, req *basepb.RegisterVehicleRequest) (*emptypb.Empty, error) {
	id := ctx.Value(internal.CtxKeyId)
	email := ctx.Value(internal.CtxKeyEmail)
	name := ctx.Value(internal.CtxKeyName)
	config := base.Config{
		UserInfo: internal.UserInfo{
			Id:    id.(string),
			Email: email.(string),
			Name:  name.(string),
		},
		Config:     s.Config,
		PrivateKey: s.PrivateKey,
	}

	return base.New((*base.Config)(&config)).RegisterVehicle(ctx, req)
}
