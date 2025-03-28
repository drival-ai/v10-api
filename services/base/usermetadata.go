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
	fmt.Fprintf(&q, "update usersmetadata set ")
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

func (s *svc) GetUserMetadata(ctx context.Context, in *base.GetUserMetadataRequest) (*base.GetUserMetadataResponse, error) {
	b, _ := json.Marshal(in)
	glog.Infof("GetUserMetadata input=%v", string(b))

	if in.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user id is empty")
	}

	var q strings.Builder
	fmt.Fprintf(&q, "select rank, points ")
	fmt.Fprintf(&q, "from usersmetadata ")
	fmt.Fprintf(&q, "where id = @id")
	args := pgx.NamedArgs{
		"id": s.Config.UserInfo.Id,
	}

	rows, err := global.PgxPool.Query(ctx, q.String(), args)
	if err != nil {
		glog.Errorf("Query failed: %v", err)
		return nil, internal.InternalErr
	}
	defer rows.Close()

	var rank string
	var points int32
	if rows.Next() {
		err = rows.Scan(&rank, &points)
		if err != nil {
			glog.Errorf("Scan failed: %v", err)
			return nil, internal.InternalErr
		}
	} else {
		glog.Errorf("No rows found")
		return nil, status.Errorf(codes.NotFound, "user metadata not found")
	}

	res := &base.GetUserMetadataResponse{
		UserMetadata: &base.UserMetadata{
			Id:     s.Config.UserInfo.Id,
			Rank:   rank,
			Points: points,
		},
	}

	glog.Info("GetUserMetadata success!")
	return res, nil
}
