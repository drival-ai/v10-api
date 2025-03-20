package main

import (
	"context"
	"sync"

	"github.com/drival-ai/v10-api/global"
	"github.com/drival-ai/v10-api/internal"
	"github.com/drival-ai/v10-go/iam/v1"

	b "github.com/drival-ai/v10-api/services/base"
	base "github.com/drival-ai/v10-go/base/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	iam.UnimplementedIamServer
	base.UnimplementedV10Server
}

func (s *service) Login(req *iam.LoginRequest, stream iam.Iam_LoginServer) error {
	return status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (s *service) WhoAmI(ctx context.Context, req *iam.WhoAmIRequest) (*iam.WhoAmIResponse, error) {
	return &iam.WhoAmIResponse{Name: "V10 MVP"}, nil
}

func (s *service) RegisterVehicle(ctx context.Context, req *base.RegisterVehicleRequest) (*emptypb.Empty, error) {
	return b.New().RegisterVehicle(ctx, req)
}
