package global

import (
	"cloud.google.com/go/spanner"
)

var (
	Client *spanner.Client // connection to Spanner
)
