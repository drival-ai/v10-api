package iam

import (
	"context"

	"github.com/drival-ai/v10-api/internal"
	"github.com/drival-ai/v10-go/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type svc struct {
	UserInfo internal.UserInfo
}

func (s *svc) Register(ctx context.Context, req *iam.RegisterRequest) (*iam.RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}

func (s *svc) Login(ctx context.Context, req *iam.LoginRequest) (*iam.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (s *svc) WhoAmI(ctx context.Context, req *iam.WhoAmIRequest) (*iam.WhoAmIResponse, error) {
	return &iam.WhoAmIResponse{Name: "V10 MVP"}, nil
}

func New() *svc {
	return &svc{}
}
