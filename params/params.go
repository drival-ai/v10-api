package params

import (
	"flag"
)

var (
	PostgresDsn = flag.String("pgdsn", "", "Postgres DSN (Data Source Name)")
)
