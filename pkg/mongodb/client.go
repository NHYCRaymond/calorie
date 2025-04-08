// Package mongodb provides a MongoDB client wrapper with connection pooling and authentication support.
// It simplifies MongoDB operations with a clean and consistent API.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config MongoDB 配置
type Config struct {
	URI      string
	Database string
	Username string
	Password string
}

// Client MongoDB 客户端
type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewClient 创建新的 MongoDB 客户端
func NewClient(config *Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.URI)
	if config.Username != "" && config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: config.Username,
			Password: config.Password,
		})
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:   client,
		database: client.Database(config.Database),
	}, nil
}

// Collection 获取集合
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Close 关闭连接
func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return c.client.Disconnect(ctx)
}
