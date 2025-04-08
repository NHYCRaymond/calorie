package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Config Redis 配置
type Config struct {
	Addr     string
	Password string
	DB       int
}

// Client Redis 客户端
type Client struct {
	client *redis.Client
}

// NewClient 创建新的 Redis 客户端
func NewClient(config *Config) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// 检查连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set 设置值
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.client.Close()
}
