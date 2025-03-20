package params

import (
	"flag"
)

var (
	ConfigFile = flag.String("config", "/etc/v10-api/config", "Config file")
)
