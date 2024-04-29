package repository

import (
	"context"

	"github.com/RafalSalwa/auth-api/cmd/auth_service/config"
	"github.com/RafalSalwa/auth-api/pkg/mongo"
	redisClient "github.com/RafalSalwa/auth-api/pkg/redis"
	"github.com/RafalSalwa/auth-api/pkg/sql"
)

const (
	MySQL   string = "mysql"
	MongoDB string = "mongo"
	Redis   string = "redis"
)

func NewUserRepository(ctx context.Context, dbType string, params *config.Config) (UserRepository, error) {
	switch dbType {
	case MySQL:
		con, err := sql.NewGormConnection(params.MySQL)
		if err != nil {
			return nil, err
		}
		return newMySQLUserRepository(con), nil

	case MongoDB:
		mongoClient, err := mongo.NewClient(ctx, params.Mongo)
		if err != nil {
			return nil, err
		}
		return newMongoDBUserRepository(mongoClient, params.Mongo), nil

	case Redis:
		universalRedisClient, err := redisClient.NewUniversalRedisClient(ctx, params.Redis)
		if err != nil {
			return nil, err
		}

		return newRedisUserRepository(universalRedisClient), nil
	default:
		panic("Unsupported database type")
	}
}
