package iam

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/drival-ai/v10-api/global"
	"github.com/drival-ai/v10-api/internal"
	"github.com/drival-ai/v10-go/iam/v1"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	glog.Infof("claims=%v", payload.Claims)

	var sub string
	if v, ok := payload.Claims["sub"]; ok {
		sub = fmt.Sprintf("%v", v)
	}

	var email string
	if v, ok := payload.Claims["email"]; ok {
		email = fmt.Sprintf("%v", v)
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
	}

	var givenName string
	if v, ok := payload.Claims["given_name"]; ok {
		givenName = fmt.Sprintf("%v", v)
	}

	var fullName string
	if v, ok := payload.Claims["name"]; ok {
		fullName = fmt.Sprintf("%v", v)
	}

	var picture string
	if v, ok := payload.Claims["picture"]; ok {
		picture = fmt.Sprintf("%v", v)
	}

	var aud string
	if v, ok := payload.Claims["aud"]; ok {
		aud = fmt.Sprintf("%v", v)
	}

	var azp string
	if v, ok := payload.Claims["azp"]; ok {
		azp = fmt.Sprintf("%v", v)
	}

	// See if already registered.
	var found bool
	var qId string
	var q strings.Builder
	fmt.Fprintf(&q, "select id from users ")
	fmt.Fprintf(&q, "where id = $1 ")
	rows, _ := global.PgxPool.Query(ctx, q.String(), sub)
	pgx.ForEachRow(rows, []any{&qId}, func() error {
		found = true
		return nil
	})

	// Add to db.
	if !found {
		q.Reset()
		fmt.Fprintf(&q, "insert into users (id, email, email_verified, ")
		fmt.Fprintf(&q, "family_name, given_name, full_name, picture, ")
		fmt.Fprintf(&q, "aud, azp) values (@id, @email, @verified, ")
		fmt.Fprintf(&q, "@family, @given, @full, @pic, @aud, @azp)")
		args := pgx.NamedArgs{
			"id":       sub,
			"email":    email,
			"verified": emailVerified,
			"family":   familyName,
			"given":    givenName,
			"full":     fullName,
			"pic":      picture,
			"aud":      aud,
			"azp":      azp,
		}

		_, err = global.PgxPool.Exec(ctx, q.String(), args)
		if err != nil {
			glog.Errorf("Exec failed: %v", err)
			return nil, internal.InternalErr
		}
	}

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
