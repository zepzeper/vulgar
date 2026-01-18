package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps a Redis client
type Client struct {
	client *redis.Client
}

// ConnectOptions holds connection configuration
type ConnectOptions struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewClient creates a new Redis client from a connection string
func NewClient(connStr string) (*Client, error) {
	opts, err := redis.ParseURL(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Client{client: client}, nil
}

// NewClientFromOptions creates a new Redis client from options
func NewClientFromOptions(opts ConnectOptions) (*Client, error) {
	// Defaults
	if opts.Host == "" {
		opts.Host = "localhost"
	}
	if opts.Port == 0 {
		opts.Port = 6379
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       opts.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Client{client: client}, nil
}

// Close closes the client
func (c *Client) Close() error {
	return c.client.Close()
}

// Get Wrappers

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Del(ctx, keys...).Result()
}

func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.client.Expire(ctx, key, expiration).Result()
}

func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// Hash

func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.HSet(ctx, key, values...).Result()
}

func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// List

func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.LPush(ctx, key, values...).Result()
}

func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.RPush(ctx, key, values...).Result()
}

func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return c.client.LPop(ctx, key).Result()
}

func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	return c.client.RPop(ctx, key).Result()
}

func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// Set

func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.client.SAdd(ctx, key, members...).Result()
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// PubSub

func (c *Client) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return c.client.Publish(ctx, channel, message).Result()
}

// Incr

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Helper to expose internal client if needed (for more advanced usage not wrapped yet)
func (c *Client) SDK() *redis.Client {
	return c.client
}

// IsNilError checks if the error is redis.Nil
func (c *Client) IsNilError(err error) bool {
	return err == redis.Nil
}
