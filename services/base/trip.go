package base

import (
	"context"
	"database/sql"
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

type Trip struct {
	Id          sql.NullString
	Vin         sql.NullString
	UserId      sql.NullString
	StartTime   sql.NullString
	EndTime     sql.NullString
	Distance    sql.NullFloat64
	Points      sql.NullInt32
	MapSnapshot sql.NullString
}

func (s *svc) StartTrip(ctx context.Context, in *base.StartTripRequest) (*base.StartTripResponse, error) {
	b, _ := json.Marshal(in)
	glog.Infof("StartTrip input=%v", string(b))

	if in.Vin == "" {
		return nil, status.Errorf(codes.InvalidArgument, "vin is empty")
	}

	if in.StartTime == "" {
		return nil, status.Errorf(codes.InvalidArgument, "start time is empty")
	}

	tripId := uuid.New().String()
	var q strings.Builder
	fmt.Fprintf(&q, "insert into trips (id, vin, ")
	fmt.Fprintf(&q, "user_id, start_time, end_time, distance) ")
	fmt.Fprintf(&q, "values (@id, @vin, ")
	fmt.Fprintf(&q, "@user_id, @start_time, @end_time, @distance)")
	args := pgx.NamedArgs{
		"id":         tripId,
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
	return &base.StartTripResponse{Id: tripId}, nil
}

func (s *svc) UpdateTrip(ctx context.Context, in *base.UpdateTripRequest) (*emptypb.Empty, error) {
	b, _ := json.Marshal(in)
	glog.Infof("UpdateTrip input=%v", string(b))
	if in.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "trip id is empty")
	}

	var q strings.Builder
	fmt.Fprintf(&q, "update trips set points = @points, ")
	fmt.Fprintf(&q, "distance = @distance, end_time = @end_time, map_snapshot = @map_snapshot  where id = @id and user_id = @user_id")
	args := pgx.NamedArgs{
		"id":           in.Id,
		"user_id":      s.Config.UserInfo.Id,
		"points":       in.Trip.Points,
		"distance":     in.Trip.Distance,
		"end_time":     in.Trip.EndTime,
		"map_snapshot": in.Trip.MapSnapshot,
	}
	_, err := global.PgxPool.Exec(ctx, q.String(), args)
	if err != nil {
		glog.Errorf("Exec failed: %v", err)
		return nil, internal.InternalErr
	}

	glog.Info("UpdateTrip success!")
	return &emptypb.Empty{}, nil
}

func (s *svc) EndTrip(ctx context.Context, in *base.EndTripRequest) (*emptypb.Empty, error) {
	b, _ := json.Marshal(in)
	glog.Infof("EndTrip input=%v", string(b))
	if in.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "trip id is empty")
	}
	if in.EndTime == "" {
		return nil, status.Errorf(codes.InvalidArgument, "end time is empty")
	}
	if in.Distance < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "distance could not be negative")
	}
	if in.MapSnapshot == "" {
		return nil, status.Errorf(codes.InvalidArgument, "map snapshot is empty")
	}

	var q strings.Builder
	fmt.Fprintf(&q, "update trips set end_time = @end_time, ")
	fmt.Fprintf(&q, "distance = @distance, points = @points, map_snapshot = @map_snapshot where id = @id and user_id = @user_id")
	args := pgx.NamedArgs{
		"id":           in.Id,
		"end_time":     in.EndTime,
		"distance":     in.Distance,
		"points":       in.Points,
		"map_snapshot": in.MapSnapshot,
		"user_id":      s.Config.UserInfo.Id,
	}
	_, err := global.PgxPool.Exec(ctx, q.String(), args)
	if err != nil {
		glog.Errorf("Exec failed: %v", err)
		return nil, internal.InternalErr
	}

	glog.Info("EndTrip success!")
	return &emptypb.Empty{}, nil
}

func (s *svc) ListTrips(in *base.ListTripsRequest, stream base.V10_ListTripsServer) error {
	ctx := stream.Context()
	var q strings.Builder
	fmt.Fprintf(&q, "select id, ")
	fmt.Fprintf(&q, "vin, start_time, end_time, distance, points, map_snapshot ")
	fmt.Fprintf(&q, "from trips ")
	fmt.Fprintf(&q, "where user_id = $1")
	rows, err := global.PgxPool.Query(ctx, q.String(), s.Config.UserInfo.Id)
	if err != nil {
		glog.Errorf("Query failed: %v", err)
		return internal.InternalErr
	}
	defer rows.Close()

	for rows.Next() {
		var v Trip
		err = rows.Scan(&v.Id,
			&v.Vin, &v.StartTime, &v.EndTime, &v.Distance, &v.Points, &v.MapSnapshot)
		if err != nil {
			glog.Errorf("Scan failed: %v", err)
			return internal.InternalErr
		}

		if err = stream.Send(&base.Trip{
			Id:          v.Id.String,
			Vin:         v.Vin.String,
			StartTime:   v.StartTime.String,
			EndTime:     v.EndTime.String,
			Distance:    float32(v.Distance.Float64),
			Points:      v.Points.Int32,
			MapSnapshot: v.MapSnapshot.String,
		}); err != nil {
			glog.Errorf("Send failed: %v", err)
			return internal.InternalErr
		}
	}

	if err = rows.Err(); err != nil {
		glog.Errorf("rows.Err failed: %v", err)
		return internal.InternalErr
	}

	glog.Infof("ListTrips success!")
	return nil
}
