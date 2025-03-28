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
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *svc) UpdateUserMetadata(ctx context.Context, in *base.UpdateUserMetadataRequest) (*emptypb.Empty, error) {
	b, _ := json.Marshal(in)
	glog.Infof("UpdateUserMetadata input=%v", string(b))

	if in.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user id is empty")
	}

	var q strings.Builder
	fmt.Fprintf(&q, "update usermetadata set ")
	fmt.Fprintf(&q, "rank = @rank, points = @points ")
	fmt.Fprintf(&q, "where id = @id")
	args := pgx.NamedArgs{
		"rank":   in.UserMetadata.Rank,
		"points": in.UserMetadata.Points,
		"id":     s.Config.UserInfo.Id,
	}

	_, err := global.PgxPool.Exec(ctx, q.String(), args)
	if err != nil {
		glog.Errorf("Exec failed: %v", err)
		return nil, internal.InternalErr
	}

	glog.Info("UpdateUserMetadata success!")
	return &emptypb.Empty{}, nil
}
