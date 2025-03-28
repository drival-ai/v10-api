package base

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/drival-ai/v10-api/global"
	"github.com/drival-ai/v10-api/internal"
	"github.com/drival-ai/v10-go/base/v1"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *svc) StartTrip(ctx context.Context, in *base.StartTripRequest) (*emptypb.Empty, error) {
	b, _ := json.Marshal(in)
	glog.Infof("StartTrip input=%v", string(b))

	if in.Vin == "" {
		return nil, status.Errorf(codes.InvalidArgument, "vin is empty")
	}

	if in.StartTime == "" {
		return nil, status.Errorf(codes.InvalidArgument, "start time is empty")
	}

	var q strings.Builder
	fmt.Fprintf(&q, "insert into trips (id, vin, ")
	fmt.Fprintf(&q, "user_id, start_time, end_time, distance) ")
	fmt.Fprintf(&q, "values (@id, @vin, ")
	fmt.Fprintf(&q, "@user_id, @start_time, @end_time, @distance)")
	args := pgx.NamedArgs{
		"id":         uuid.New().String(),
		"vin":        in.Vin,
		"user_id":    s.Config.UserInfo.Id,
		"start_time": in.StartTime,
		"end_time":   "",
		"distance":   0,
	}

	_, err := global.PgxPool.Exec(ctx, q.String(), args)
	if err != nil {
		glog.Errorf("Exec failed: %v", err)
		return nil, internal.InternalErr
	}

	glog.Info("StartTrip success!")
	return &emptypb.Empty{}, nil
}
