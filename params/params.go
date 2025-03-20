package params

import (
	"flag"
)

var (
	PostgresDsn = flag.String("pg", "", "Postgres DSN (Data Source Name)")
)
