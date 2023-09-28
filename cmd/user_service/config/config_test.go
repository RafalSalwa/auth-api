//go:build unit

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	path, err := getEnvPath()
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "user_service")

	c, err := InitConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, c.ServiceName)
	assert.NotEmpty(t, c.App)
	assert.NotEmpty(t, c.Redis)
	assert.NotEmpty(t, c.MySQL)
	assert.NotEmpty(t, c.GRPC)
	assert.NotEmpty(t, c.Mongo)
	assert.NotEmpty(t, c.JWTToken)
	assert.NotEmpty(t, c.Jaeger)
	assert.NotEmpty(t, c.Logger)
	assert.Equal(t, "user_service", c.ServiceName)
}
