package global

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	AuthPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA331E8VogGVAwiIMlkxnM
DQ2DY/Uvf/MevmI4E1Hd0JciNVLQjgmmrSqiHcKSXa1Zs8Wls3bytc2Sg/WIvebf
40LJ0jnxVsOKQlFJ25QgHEzaldlxrJPHEcMc/3+bxVkiWirSddwP2561YOOygF7q
55r3vRDmO6d3RYyI5I+lynGvrPXdcSyiyjT09FZDRcr/grsBiAU1gDfhQHc3SlzG
A0tG52vkv3DeE1xjOQK4PEyMqy0noRHleTaY3bZ0IOMAqHEkhU1wiQXMbkepytzN
HNmLYtH+fTcM7JqcfbhBl2QJ5w6/oSZueA/ugNpX14DJIFvl3Ux+F62zJvZvXKC8
oQIDAQAB
-----END PUBLIC KEY-----`
)

var (
	PgxPool *pgxpool.Pool
)

type Config struct {
	AndroidClientId string `yaml:"android-client-id"` // used for audience
	PgDsn           string `yaml:"pg-dsn"`
}

func LoadPublicKey() (*rsa.PublicKey, error) {
	data, _ := pem.Decode([]byte(AuthPublicKey))
	pub, err := x509.ParsePKIXPublicKey(data.Bytes)
	if err != nil {
		return nil, err
	}

	pubk := pub.(*rsa.PublicKey)
	return pubk, err
}
