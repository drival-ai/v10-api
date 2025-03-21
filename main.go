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
	jwtv5 "github.com/golang-jwt/jwt/v5"
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

	// Setup private key:
	pkb, err := os.ReadFile(*params.PrivateKey)
	if err != nil {
		glog.Fatalf("ReadFile (%v) failed: %v", *params.PrivateKey, err)
	}

	pk, err := jwtv5.ParseRSAPrivateKeyFromPEM(pkb)
	if err != nil {
		glog.Fatalf("ParseRSAPrivateKeyFromPEM failed: %v", err)
	}

	// currentTime := time.Now()
	// atClaims := jwtv5.MapClaims{}
	// atClaims["iss"] = "app.alphaus.cloud"
	// atClaims["aud"] = "CloudSaverTagManager"
	// atClaims["jti"] = uuid.NewString()
	// atClaims["iat"] = currentTime.Unix()
	// atClaims["nbf"] = currentTime.Unix()
	// atClaims["exp"] = currentTime.Add(time.Second * 259200).Unix() // 1d=86400
	// atClaims["sub"] = "userid1"
	// atClaims["tm_companyid"] = "MSP-5aa311904d5d6"
	// atClaims["tm_aws"] = []map[string]string{
	// 	{
	// 		"accountId":  "_120802311370_1",
	// 		"externalId": "323676f1-e768-49f1-ad59-24e0c8886177",
	// 	},
	// }

	// at := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, atClaims)
	// token, err := at.SignedString(prvKey)
	// if err != nil {
	// 	logger.Errorf("SignedString failed: %v", err)
	// 	return
	// }

	// logger.Info(token)

	defer l.Close()
	auth := &internal.Auth{
		AndroidClientId: config.AndroidClientId,
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

	svc := &service{
		ctx:        ctx,
		Config:     &config,
		PrivateKey: pk,
	}

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
