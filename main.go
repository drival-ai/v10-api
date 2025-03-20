package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/spanner"
	"github.com/drival-ai/v10-api/internal"
	"github.com/drival-ai/v10-api/params"
	"github.com/drival-ai/v10-go/iam/v1"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"google.golang.org/grpc"
)

var (
	client *spanner.Client
)

func run(ctx context.Context, network, port string, done chan error) error {
	l, err := net.Listen(network, ":"+port)
	if err != nil {
		glog.Errorf("net.Listen failed: %v", err)
		return err
	}

	pgdsn := *params.PostgresDsn
	if pgdsn == "" {
		b, err := os.ReadFile("/etc/v10-api/postgres")
		if err != nil {
			glog.Errorf("ReadFile(/etc/v10-api/postgres) failed: %v", err)
		} else {
			pgdsn = string(b)
		}
	}

	glog.Infof("pg=%v", pgdsn)

	defer l.Close()
	internalData := &internal.InternalData{
		RunEnv:   "",
		Audience: "https://",
	}

	// Setup our grpc server.
	gs := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			ratelimit.UnaryServerInterceptor(&limiter{}),
			grpc.UnaryServerInterceptor(internalData.UnaryInterceptor),
		),
		grpc.ChainStreamInterceptor(
			ratelimit.StreamServerInterceptor(&limiter{}),
			grpc.StreamServerInterceptor(internalData.StreamInterceptor),
		),
	)

	svc := &service{ctx: ctx, client: client}
	iam.RegisterIamServer(gs, svc)

	go func() {
		<-ctx.Done()
		gs.GracefulStop()
		done <- nil
	}()

	return gs.Serve(l)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("todo")
		}
	}()

	flag.Parse()
	defer glog.Flush()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
			glog.Infof("defaulting to port %s", port)
		}

		glog.Infof("serving grpc at :%v", port)
		run(ctx, "tcp", port, done)
	}()

	// Interrupt handler.
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		glog.Infof("signal: %v", <-sigch)
		cancel()
	}()

	<-done
}
