package redis

import (
	redis "gopkg.in/redis.v3"
)

var client *Client

var Nil = redis.Nil

// Config is the redis configuration
type Config struct {
	Addr     string
	DB       int
	PoolSize int // defaults to 10
}

// Client is a wrapper around redis.Client
type Client struct {
	*redis.Client
}

// Init initializes the redis client
func Init(c *Config) error {
	client = &Client{redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		DB:       int64(c.DB),
		PoolSize: c.PoolSize,
	})}
	status := client.Ping()
	return status.Err()
}

// GetClient returns the initialized client
func GetClient() *Client {
	return client
}
