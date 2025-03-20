package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/drival-ai/v10-api/global"
	"github.com/drival-ai/v10-api/internal"
	"github.com/drival-ai/v10-api/params"
	"github.com/drival-ai/v10-go/base/v1"
	"github.com/drival-ai/v10-go/iam/v1"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v3"
)

func run(ctx context.Context, network, port string, done chan error) error {
	l, err := net.Listen(network, ":"+port)
	if err != nil {
		glog.Errorf("net.Listen failed: %v", err)
		return err
	}

	var config global.Config
	if *params.ConfigFile != "" {
		b, err := os.ReadFile(*params.ConfigFile)
		if err != nil {
			glog.Fatalf("ReadFile(%v) failed: %v", *params.ConfigFile, err)
		} else {
			err := yaml.Unmarshal(b, &config)
			if err != nil {
				glog.Fatalf("Unmarshal failed: %v", err)
			}
		}
	}

	// Test connection to RDS/Postgres:
	global.PgxPool, err = pgxpool.New(ctx, config.PgDsn)
	if err != nil {
		glog.Errorf("pgxpool.New failed: %v", err)
	} else {
		err = global.PgxPool.Ping(ctx)
		if err != nil {
			glog.Errorf("Ping failed: %v", err)
		} else {
			glog.Info("ping ok")
		}
	}

	defer l.Close()
	auth := &internal.Auth{
		Audience: "https://",
	}

	// Setup our grpc server.
	gs := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			ratelimit.UnaryServerInterceptor(&limiter{}),
			grpc.UnaryServerInterceptor(auth.UnaryInterceptor),
		),
		grpc.ChainStreamInterceptor(
			ratelimit.StreamServerInterceptor(&limiter{}),
			grpc.StreamServerInterceptor(auth.StreamInterceptor),
		),
	)

	svc := &service{ctx: ctx, Config: &config}
	iam.RegisterIamServer(gs, svc)
	base.RegisterV10Server(gs, svc)

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
