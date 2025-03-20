package internal

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	CtxKeyEmail      = "email"
	CtxKeyFullMethod = "fullMethod"
)

var (
	unauthorizedCallerErr = status.Errorf(codes.Unauthenticated, "Unauthorized caller.")

	allowed = []string{
		"@labs-169405.iam.gserviceaccount.com",  // dev
		"@mobingi-main.iam.gserviceaccount.com", // next, prod
	}
)

type UserInfo struct {
	Email string
}

type InternalData struct {
	RunEnv   string // dev,next, prod
	Audience string // audience for token validation (CloudRun)
}

func (d *InternalData) verifyCaller(ctx context.Context, md metadata.MD) (UserInfo, error) {
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
		return UserInfo{}, unauthorizedCallerErr
	}

	payload, err := idtoken.Validate(ctx, token, d.Audience)
	if err != nil {
		glog.Errorf("Validate failed: %v", err)
		return UserInfo{}, err
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
		return UserInfo{}, unauthorizedCallerErr
	}

	var email string
	if v, ok := payload.Claims["email"]; ok {
		email = fmt.Sprintf("%v", v)
	}

	var validEmail bool
	for _, allow := range allowed {
		if strings.HasSuffix(email, allow) {
			validEmail = validEmail || true
		}
	}

	if !validEmail {
		glog.Errorf("failed: invalid email")
		return UserInfo{}, unauthorizedCallerErr
	}

	return UserInfo{Email: email}, nil
}

func (d *InternalData) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (interface{}, error) {
	defer func(begin time.Time) {
		glog.Infof("[unary] << %v duration: %v", info.FullMethod, time.Since(begin))
	}(time.Now())

	glog.Infof("[unary] >> %v", info.FullMethod)
	md, _ := metadata.FromIncomingContext(ctx)
	u, err := d.verifyCaller(ctx, md)
	if err != nil {
		return nil, unauthorizedCallerErr
	}

	nctx := metadata.NewIncomingContext(ctx, md)
	nctx = context.WithValue(nctx, CtxKeyEmail, u.Email)
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

func (d *InternalData) StreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, h grpc.StreamHandler) error {
	defer func(begin time.Time) {
		glog.Infof("[stream] << %v duration: %v", info.FullMethod, time.Since(begin))
	}(time.Now())

	glog.Infof("[stream] >> %v", info.FullMethod)
	md, _ := metadata.FromIncomingContext(stream.Context())
	u, err := d.verifyCaller(stream.Context(), md)
	if err != nil {
		return unauthorizedCallerErr
	}

	wrap := newStreamContextWrapper(stream)
	nctx := context.WithValue(wrap.Context(), CtxKeyEmail, u.Email)
	nctx = context.WithValue(nctx, CtxKeyFullMethod, info.FullMethod)
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
