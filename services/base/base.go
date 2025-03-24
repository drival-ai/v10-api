package base

import (
	"context"
	"crypto/rsa"
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

type Config struct {
	UserInfo   internal.UserInfo
	Config     *global.Config
	PrivateKey *rsa.PrivateKey
}

type svc struct {
	Config *Config
}

type Vehicle struct {
	Id            sql.NullString
	ChassisNumber sql.NullString
	Vin           sql.NullString
	Make          sql.NullString
	Model         sql.NullString
	Year          sql.NullInt64
	Kilometers    sql.NullInt64
}

func (s *svc) RegisterVehicle(ctx context.Context, in *base.RegisterVehicleRequest) (*emptypb.Empty, error) {
	b, _ := json.Marshal(in)
	glog.Infof("RegisterVehicle input=%v", string(b))
	if in.Vehicle == nil {
		return nil, status.Errorf(codes.InvalidArgument, "vehicle is nil")
	}
	// Check if vehicle already exists
	var exist bool
	if in.Vehicle.Vin != "" {
		err := global.PgxPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM vehicles WHERE vin = $1)", in.Vehicle.Vin).Scan(&exist)
		if err != nil {
			glog.Errorf("QueryRow failed: %v", err)
			return nil, internal.InternalErr
		}
		if exist {
			return nil, status.Errorf(codes.AlreadyExists, "vehicle already exists")
		}
	} else if in.Vehicle.ChassisNumber != "" {
		err := global.PgxPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM vehicles WHERE chassis_number = $1)", in.Vehicle.ChassisNumber).Scan(&exist)
		if err != nil {
			glog.Errorf("QueryRow failed: %v", err)
			return nil, internal.InternalErr
		}
		if exist {
			return nil, status.Errorf(codes.AlreadyExists, "vehicle already exists")
		}
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "vin and chassis number are empty")
	}

	vehicleId := uuid.New().String()
	var q strings.Builder
	fmt.Fprintf(&q, "insert into vehicles (id, chassis_number, vin, ")
	fmt.Fprintf(&q, "make, model, year, kms, user_id) ")
	fmt.Fprintf(&q, "values (@id, @chassis_number, @vin, ")
	fmt.Fprintf(&q, "@make, @model, @year, @kms, @user_id)")
	args := pgx.NamedArgs{
		"id":             vehicleId,
		"chassis_number": in.Vehicle.ChassisNumber,
		"make":           in.Vehicle.Make,
		"model":          in.Vehicle.Model,
		"year":           in.Vehicle.Year,
		"kms":            in.Vehicle.Kilometers,
		"user_id":        s.Config.UserInfo.Id,
	}

	_, err := global.PgxPool.Exec(ctx, q.String(), args)
	if err != nil {
		glog.Errorf("Exec failed: %v", err)
		return nil, internal.InternalErr
	}

	glog.Info("RegisterVehicle success!")
	return &emptypb.Empty{}, nil
}

func (s *svc) ListVehicles(ctx context.Context, in *base.ListVehiclesRequest) (*base.ListVehiclesResponse, error) {
	var q strings.Builder
	fmt.Fprintf(&q, "select id, chassis_number, vin, ")
	fmt.Fprintf(&q, "make, model, year, kms ")
	fmt.Fprintf(&q, "from vehicles ")
	fmt.Fprintf(&q, "where user_id = $1")
	rows, err := global.PgxPool.Query(ctx, q.String(), s.Config.UserInfo.Id)
	if err != nil {
		glog.Errorf("Query failed: %v", err)
		return nil, internal.InternalErr
	}
	defer rows.Close()

	var vehicles []*base.Vehicle
	for rows.Next() {
		var v Vehicle
		err = rows.Scan(&v.Id, &v.ChassisNumber, &v.Vin,
			&v.Make, &v.Model, &v.Year, &v.Kilometers)
		if err != nil {
			glog.Errorf("Scan failed: %v", err)
			return nil, internal.InternalErr
		}
		vehicles = append(vehicles, &base.Vehicle{
			Id:            v.Id.String,
			ChassisNumber: v.ChassisNumber.String,
			Vin:           v.Vin.String,
			Make:          v.Make.String,
			Model:         v.Model.String,
			Year:          int32(v.Year.Int64),
			Kilometers:    int32(v.Kilometers.Int64),
		})
	}

	if err = rows.Err(); err != nil {
		glog.Errorf("rows.Err failed: %v", err)
		return nil, internal.InternalErr
	}

	glog.Infof("ListVehicles success!")
	return &base.ListVehiclesResponse{Vehicles: vehicles}, nil
}

func (s *svc) DeleteVehicle(ctx context.Context, in *base.DeleteVehicleRequest) (*emptypb.Empty, error) {
	b, _ := json.Marshal(in)
	glog.Infof("DeleteVehicle input=%v", string(b))
	if in.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id is empty")
	}

	var q strings.Builder
	fmt.Fprintf(&q, "delete from vehicles where id = $1 and user_id = $2")
	_, err := global.PgxPool.Exec(ctx, q.String(), in.Id, s.Config.UserInfo.Id)
	if err != nil {
		glog.Errorf("Exec failed: %v", err)
		return nil, internal.InternalErr
	}

	glog.Info("DeleteVehicle success!")
	return &emptypb.Empty{}, nil
}

func New(config *Config) *svc { return &svc{Config: config} }
