package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/drival-ai/v10-api/global"
	"github.com/drival-ai/v10-api/internal"
	"github.com/drival-ai/v10-go/iam/v1"
	"github.com/golang/glog"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	UserInfo internal.UserInfo
	Config   *global.Config
}

type svc struct {
	Config *Config
}

func (s *svc) Register(ctx context.Context, req *iam.RegisterRequest) (*iam.RegisterResponse, error) {
	if req == nil {
		return nil, internal.UnauthorizedCallerErr
	}

	if req.Token == "" {
		glog.Errorf("failed: empty token")
		return nil, internal.UnauthorizedCallerErr
	}

	payload, err := idtoken.Validate(ctx, req.Token, s.Config.Config.AndroidClientId)
	if err != nil {
		glog.Errorf("Validate failed: %v", err)
		return nil, internal.UnauthorizedCallerErr
	}

	b, _ := json.Marshal(payload)
	glog.Infof("payload=%v", string(b))
	glog.Infof("claims=%v", payload.Claims)

	var emailVerified bool
	if v, ok := payload.Claims["email_verified"]; ok {
		emailVerified = v.(bool)
	}

	if !emailVerified {
		glog.Errorf("failed: email not verified")
		return nil, internal.UnauthorizedCallerErr
	}

	var email string
	if v, ok := payload.Claims["email"]; ok {
		email = fmt.Sprintf("%v", v)
		glog.Infof("email=%v", email)
	}

	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}

func (s *svc) Login(ctx context.Context, req *iam.LoginRequest) (*iam.LoginResponse, error) {
	if req == nil {
		return nil, internal.UnauthorizedCallerErr
	}

	if req.Token == "" {
		glog.Errorf("failed: empty token")
		return nil, internal.UnauthorizedCallerErr
	}

	payload, err := idtoken.Validate(ctx, req.Token, s.Config.Config.AndroidClientId)
	if err != nil {
		glog.Errorf("Validate failed: %v", err)
		return nil, internal.UnauthorizedCallerErr
	}

	b, _ := json.Marshal(payload)
	glog.Infof("payload=%v", string(b))
	glog.Infof("claims=%v", payload.Claims)

	var emailVerified bool
	if v, ok := payload.Claims["email_verified"]; ok {
		emailVerified = v.(bool)
	}

	if !emailVerified {
		glog.Errorf("failed: email not verified")
		return nil, internal.UnauthorizedCallerErr
	}

	var email string
	if v, ok := payload.Claims["email"]; ok {
		email = fmt.Sprintf("%v", v)
		glog.Infof("email=%v", email)
	}

	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (s *svc) WhoAmI(ctx context.Context, req *iam.WhoAmIRequest) (*iam.WhoAmIResponse, error) {
	return &iam.WhoAmIResponse{Name: "V10 MVP"}, nil
}

func New(config *Config) *svc { return &svc{Config: config} }
