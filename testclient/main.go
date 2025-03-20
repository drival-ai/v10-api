package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"log/slog"
	"strings"

	"github.com/drival-ai/v10-go/iam/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	addr = flag.String("addr", "", "The server address in the format of host:port")
)

func main() {
	flag.Parse()
	ctx := context.Background()
	// dev: iamd-dev-cnugyv5cta-an.a.run.app
	// next: iamd-next-u554nqhjka-an.a.run.app
	// prod: iamd-prod-u554nqhjka-an.a.run.app
	svc := "iamd-next-u554nqhjka-an.a.run.app"
	var opts []grpc.DialOption
	if strings.Contains(*addr, "localhost") {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	// ts, err := idtoken.NewTokenSource(ctx, "https://"+svc)
	// if err != nil {
	// 	slog.Error("NewTokenSource failed:", "e", err)
	// 	return
	// }

	// token, err := ts.Token()
	// if err != nil {
	// 	slog.Error("Token failed:", "e", err)
	// 	return
	// }

	opts = append(opts, grpc.WithUnaryInterceptor(func(ctx context.Context,
		method string, req, reply any, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token.AccessToken)
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer token")
		return invoker(ctx, method, req, reply, cc, opts...)
	}))

	opts = append(opts, grpc.WithStreamInterceptor(func(ctx context.Context,
		desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token.AccessToken)
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer token")
		return streamer(ctx, desc, cc, method, opts...)
	}))

	hp := svc + ":443"
	if *addr != "" {
		hp = *addr
	}

	ccon, err := grpc.NewClient(hp, opts...)
	if err != nil {
		slog.Error("NewClient failed:", "e", err)
		return
	}

	defer ccon.Close()
	client := iam.NewIamClient(ccon)
	out, err := client.WhoAmI(ctx, &iam.WhoAmIRequest{})
	b, _ := json.Marshal(out)
	slog.Info(string(b))
}
