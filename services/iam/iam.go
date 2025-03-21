package iam

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/drival-ai/v10-api/global"
	"github.com/drival-ai/v10-api/internal"
	"github.com/drival-ai/v10-go/iam/v1"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	UserInfo   internal.UserInfo
	Config     *global.Config
	PrivateKey *rsa.PrivateKey
}

type svc struct {
	Config *Config
}

// payload={"iss":"https://accounts.google.com","aud":"1097225870985-eh9jdo7fecta4vfsgc98qcqu2kd63emf.apps.googleusercontent.com","exp":1742543624,"iat":1742540024,"sub":"109527378988750624019"}
// claims=map[aud:1097225870985-eh9jdo7fecta4vfsgc98qcqu2kd63emf.apps.googleusercontent.com azp:1097225870985-ka31pcsi8618t9sl3lseqhvs8eo85l6d.apps.googleusercontent.com email:wewpeligrino@gmail.com email_verified:true exp:1.742543624e+09 family_name:Fingerstyle given_name:N iat:1.742540024e+09 iss:https://accounts.google.com name:N Fingerstyle picture:https://lh3.googleusercontent.com/a/ACg8ocL8I3YJA-Id1bE8WYu4K1H52L15Xe5fvm1xUtK6Hh0pSE3iHA=s96-c sub:109527378988750624019]

// NOTE: Skipped by internal interceptor. Verify by ourselves.
func (s *svc) Register(ctx context.Context, req *iam.RegisterRequest) (*iam.RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}

// NOTE: Skipped by internal interceptor. Verify by ourselves.
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

	var sub string
	if v, ok := payload.Claims["sub"]; ok {
		sub = fmt.Sprintf("%v", v)
		glog.Infof("sub=%v", sub)
	}

	var email string
	if v, ok := payload.Claims["email"]; ok {
		email = fmt.Sprintf("%v", v)
		glog.Infof("email=%v", email)
	}

	var emailVerified bool
	if v, ok := payload.Claims["email_verified"]; ok {
		emailVerified = v.(bool)
	}

	if !emailVerified {
		glog.Errorf("failed: email not verified")
		return nil, internal.UnauthorizedCallerErr
	}

	var familyName string
	if v, ok := payload.Claims["family_name"]; ok {
		familyName = fmt.Sprintf("%v", v)
		glog.Infof("familyName=%v", familyName)
	}

	var givenName string
	if v, ok := payload.Claims["given_name"]; ok {
		givenName = fmt.Sprintf("%v", v)
		glog.Infof("givenName=%v", givenName)
	}

	// TODO: Save these info to users table.

	currentTime := time.Now().UTC()
	atClaims := jwtv5.MapClaims{}
	atClaims["iss"] = "app.drival.ai"
	atClaims["aud"] = "V10 Platform"
	atClaims["jti"] = uuid.NewString()
	atClaims["iat"] = currentTime.Unix()
	atClaims["nbf"] = currentTime.Unix()
	atClaims["exp"] = currentTime.Add(time.Second * 259200).Unix() // 1d=86400
	atClaims["sub"] = sub
	at := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, atClaims)
	token, err := at.SignedString(s.Config.PrivateKey)
	if err != nil {
		glog.Errorf("SignedString failed: %v", err)
		return nil, internal.UnauthorizedCallerErr
	}

	return &iam.LoginResponse{AccessToken: token}, nil
}

func (s *svc) WhoAmI(ctx context.Context, req *iam.WhoAmIRequest) (*iam.WhoAmIResponse, error) {
	return &iam.WhoAmIResponse{Name: "V10 MVP"}, nil
}

func New(config *Config) *svc { return &svc{Config: config} }
