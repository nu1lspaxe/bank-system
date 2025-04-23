package redis

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func NewRedisClient() *redis.Client {
	addr := viper.GetString("redis.addr")
	password := viper.GetString("redis.password")
	db := viper.GetInt("redis.db")

	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}
