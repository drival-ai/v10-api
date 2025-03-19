package main

import (
	"context"
	"sync"

	"cloud.google.com/go/spanner"
	"github.com/drival-ai/v10-go/iam/v1"
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
