package redis

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(config config.RedisConfig) *RedisClient {
	// Initialize Redis client here using RedisConfig
	// For example, using go-redis package:
	// rdb := redis.NewClient(&redis.Options{
	// 	Addr:     RedisConfig.RedisAddress,
	// 	Username: RedisConfig.RedisUsername,
	// 	Password: RedisConfig.RedisPassword,
	// 	DB:       RedisConfig.RedisDB,
	// })
	//
	// return &RedisClient{client: rdb}, nil

	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress,
		Username: config.RedisUsername,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	return &RedisClient{
		client: rdb,
	}

}
