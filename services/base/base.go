package base

import (
	"context"
	"encoding/json"

	"github.com/drival-ai/v10-api/internal"
	"github.com/drival-ai/v10-go/base/v1"
	"github.com/golang/glog"
	"google.golang.org/protobuf/types/known/emptypb"
)

type svc struct {
	UserInfo internal.UserInfo
}

func (s *svc) RegisterVehicle(ctx context.Context, in *base.RegisterVehicleRequest) (*emptypb.Empty, error) {
	b, _ := json.Marshal(in)
	glog.Infof("RegisterVehicle input=%v", string(b))
	return &emptypb.Empty{}, nil
}

func New() *svc {
	return &svc{}
}
