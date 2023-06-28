package csrf

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
)

type Config struct {
	salt string `mapstructure:"salt"`
}

func MakeToken(cfg Config) string {
	hash := sha256.New()
	_, err := io.WriteString(hash, cfg.salt)
	if err != nil {
	}
	token := base64.RawStdEncoding.EncodeToString(hash.Sum(nil))
	return token
}

func ValidateToken(token string, cfg Config) bool {
	trueToken := MakeToken(cfg)
	return token == trueToken
}
