package internal

import (
	"context"
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/drival-ai/v10-api/global"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/golang/glog"
	"github.com/jackc/pgx/v5"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	CtxKeyId         = "id"
	CtxKeyEmail      = "email"
	CtxKeyName       = "name"
	CtxKeyFullMethod = "fullMethod"
)

var (
	UnauthorizedCallerErr = status.Errorf(codes.Unauthenticated, "Unauthorized caller.")
	InternalErr           = status.Errorf(codes.Internal, "Internal error.")

	allowed = []string{
		"@labs-169405.iam.gserviceaccount.com",  // dev
		"@mobingi-main.iam.gserviceaccount.com", // next, prod
	}

	reBypassMethods = []string{
		`.*iam.v[0-9]*.Iam/Login.*`,
		`.*iam.v[0-9]*.Iam/Register.*`,
	}
)

type UserInfo struct {
	Id    string
	Email string
	Name  string
}

type Auth struct {
	AndroidClientId string // audience for token validation (Android)
}

func (a *Auth) verifyCaller(ctx context.Context, md metadata.MD) (UserInfo, error) {
	glog.Infof("metadata: %v", md)

	var token string
	v := md.Get("authorization")
	if len(v) > 0 {
		tt := strings.Split(v[0], " ")
		if strings.ToLower(tt[0]) == "bearer" {
			token = tt[1]
		}
	}

	if token == "" {
		glog.Errorf("failed: unauthorized call")
		return UserInfo{}, UnauthorizedCallerErr
	}

	key, err := jwtv5.ParseRSAPublicKeyFromPEM([]byte(global.AuthPublicKey))
	if err != nil {
		glog.Errorf("ParseRSAPublicKeyFromPEM failed: %v", err)
		return UserInfo{}, UnauthorizedCallerErr
	}

	claims := jwtv5.MapClaims{}
	tk, err := jwtv5.ParseWithClaims(token, &claims, func(tk *jwtv5.Token) (any, error) {
		return key, nil
	})

	_ = tk
	glog.Infof("claims=%v", claims)

	// payload, err := idtoken.Validate(ctx, token, a.AndroidClientId)
	// if err != nil {
	// 	glog.Errorf("Validate failed: %v", err)
	// 	return UserInfo{}, err
	// }

	// b, _ := json.Marshal(payload)
	// glog.Infof("payload=%v", string(b))
	// glog.Infof("claims=%v", payload.Claims)

	// var emailVerified bool
	// if v, ok := payload.Claims["email_verified"]; ok {
	// 	emailVerified = v.(bool)
	// }

	// if !emailVerified {
	// 	glog.Errorf("failed: email not verified")
	// 	return UserInfo{}, UnauthorizedCallerErr
	// }

	var sub string
	if v, ok := claims["sub"]; ok {
		sub = fmt.Sprintf("%v", v)
	}

	var found bool
	var email, name string
	var q strings.Builder
	fmt.Fprintf(&q, "select email, full_name from users ")
	fmt.Fprintf(&q, "where id = $1 ")
	rows, err := global.PgxPool.Query(ctx, q.String(), sub)
	if err != nil {
		glog.Errorf("[dbg] Query failed: %v", err)
	}

	_, err = pgx.ForEachRow(rows, []any{&email, &name}, func() error {
		found = true
		return nil
	})

	if err != nil {
		glog.Errorf("[dbg] ForEachRow failed: %v", err)
	}

	if !found {
		glog.Errorf("id %v not found", sub)
		return UserInfo{}, fmt.Errorf("not found")
	}

	return UserInfo{Id: sub, Email: email, Name: name}, nil
}

func (a *Auth) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
	defer func(begin time.Time) {
		glog.Infof("[unary] << %v duration: %v", info.FullMethod, time.Since(begin))
	}(time.Now())

	glog.Infof("[unary] >> %v", info.FullMethod)

	md, _ := metadata.FromIncomingContext(ctx)
	nctx := metadata.NewIncomingContext(ctx, md)
	if !shouldBypassMethod(info.FullMethod) {
		u, err := a.verifyCaller(ctx, md)
		if err != nil {
			return nil, UnauthorizedCallerErr
		}

		glog.Infof("user=%v", u)
		nctx = context.WithValue(nctx, CtxKeyId, u.Id)
		nctx = context.WithValue(nctx, CtxKeyEmail, u.Email)
		nctx = context.WithValue(nctx, CtxKeyName, u.Name)
	}

	nctx = context.WithValue(nctx, CtxKeyFullMethod, info.FullMethod)
	return h(nctx, req)
}

type StreamContextWrapper interface {
	grpc.ServerStream
	SetContext(context.Context)
}

type wrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrapper) Context() context.Context {
	return w.ctx
}

func (w *wrapper) SetContext(ctx context.Context) {
	w.ctx = ctx
}

func newStreamContextWrapper(inner grpc.ServerStream) StreamContextWrapper {
	ctx := inner.Context()
	return &wrapper{inner, ctx}
}

func (a *Auth) StreamInterceptor(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, h grpc.StreamHandler) error {
	defer func(begin time.Time) {
		glog.Infof("[stream] << %v duration: %v", info.FullMethod, time.Since(begin))
	}(time.Now())

	glog.Infof("[stream] >> %v", info.FullMethod)
	md, _ := metadata.FromIncomingContext(stream.Context())
	wrap := newStreamContextWrapper(stream)
	nctx := context.WithValue(wrap.Context(), CtxKeyFullMethod, info.FullMethod)
	if !shouldBypassMethod(info.FullMethod) {
		u, err := a.verifyCaller(stream.Context(), md)
		if err != nil {
			return UnauthorizedCallerErr
		}

		nctx = context.WithValue(nctx, CtxKeyId, u.Id)
		nctx = context.WithValue(nctx, CtxKeyEmail, u.Email)
		nctx = context.WithValue(nctx, CtxKeyName, u.Name)
	}

	wrap.SetContext(nctx)
	return h(srv, wrap)
}

// GetInternalConn gets a gRPC connection to an internal service.
func GetInternalConn(ctx context.Context, host string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	opts = append(opts, grpc.WithTransportCredentials(creds))
	opts = append(opts, grpc.WithBlock())
	ts, err := idtoken.NewTokenSource(ctx, "https://"+host)
	if err != nil {
		glog.Errorf("NewTokenSource failed: %v", err)
		return nil, err
	}

	token, err := ts.Token()
	if err != nil {
		glog.Errorf("Token failed: %v", err)
		return nil, err
	}

	opts = append(opts, grpc.WithUnaryInterceptor(func(ctx context.Context,
		method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token.AccessToken)
		return invoker(ctx, method, req, reply, cc, opts...)
	}))

	opts = append(opts, grpc.WithStreamInterceptor(func(ctx context.Context,
		desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token.AccessToken)
		return streamer(ctx, desc, cc, method, opts...)
	}))

	con, err := grpc.DialContext(ctx, host+":443", opts...)
	if err != nil {
		glog.Errorf("DialContext failed: %v", err)
		return nil, err
	}

	return con, nil
}

func shouldBypassMethod(method string) bool {
	var skip bool
	for _, v := range reBypassMethods {
		re := regexp.MustCompile(v)
		skip = skip || re.MatchString(method)
		if skip {
			break
		}
	}

	return skip
}
