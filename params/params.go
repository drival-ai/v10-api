package params

import (
	"flag"
)

var (
	ConfigFile = flag.String("config", "/etc/v10-api/config", "Config file")
	PrivateKey = flag.String("prvkey", "/etc/v10-api/drival.pem", "Auth private key file")
)
