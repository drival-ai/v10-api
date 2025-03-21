package global

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	PgxPool *pgxpool.Pool
)

type Config struct {
	AndroidClientId string `yaml:"android-client-id"` // used for audience
	PgDsn           string `yaml:"pg-dsn"`
}
