package main

import (
	"context"
	"sync"

	"cloud.google.com/go/spanner"
	"github.com/drival-ai/v10-go/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	cctx = func(p context.Context) context.Context {
		return context.WithValue(p, struct{}{}, nil)
	}
)

type service struct {
	ctx        context.Context
	client     *spanner.Client
	clientOnce sync.Once

	iam.UnimplementedIamServer
}

func (s *service) Login(req *iam.LoginRequest, stream iam.Iam_LoginServer) error {
	return status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (s *service) WhoAmI(ctx context.Context, req *iam.WhoAmIRequest) (*iam.WhoAmIResponse, error) {
	return &iam.WhoAmIResponse{Name: "V10 MVP"}, nil
}
