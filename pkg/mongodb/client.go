// Package mongodb provides a MongoDB client wrapper with connection pooling and authentication support.
// It simplifies MongoDB operations with a clean and consistent API.
package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/NHYCRaymond/calorie/pkg/metrics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config MongoDB 配置
type Config struct {
	URI            string
	Database       string
	Username       string
	Password       string
	MaxPoolSize    uint64
	MinPoolSize    uint64
	ConnectTimeout time.Duration
	SocketTimeout  time.Duration
	// 是否启用指标收集
	EnableMetrics bool
	// 服务名称
	ServiceName string
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	MaxPoolSize:    100,
	MinPoolSize:    10,
	ConnectTimeout: 10 * time.Second,
	SocketTimeout:  30 * time.Second,
	EnableMetrics:  true,
	ServiceName:    "mongodb",
}

// Client MongoDB 客户端
type Client struct {
	client   *mongo.Client
	database *mongo.Database
	config   *Config
	metrics  *metrics.Client
}

// NewClient 创建新的 MongoDB 客户端
func NewClient(config *Config, metricsClient *metrics.Client) (*Client, error) {
	if config == nil {
		config = DefaultConfig
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.URI)
	if config.Username != "" && config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: config.Username,
			Password: config.Password,
		})
	}

	// 设置连接池
	clientOptions.SetMaxPoolSize(config.MaxPoolSize)
	clientOptions.SetMinPoolSize(config.MinPoolSize)
	clientOptions.SetSocketTimeout(config.SocketTimeout)

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
		config:   config,
		metrics:  metricsClient,
	}, nil
}

// Collection 获取集合
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Close 关闭连接
func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.ConnectTimeout)
	defer cancel()
	return c.client.Disconnect(ctx)
}

// WithTransaction 执行事务
func (c *Client) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	session, err := c.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, fn)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// InsertOne 插入单个文档
func (c *Client) InsertOne(ctx context.Context, collection string, document interface{}) (*mongo.InsertOneResult, error) {
	start := time.Now()
	result, err := c.Collection(collection).InsertOne(ctx, document)
	c.recordMetrics("insert_one", collection, err, start)
	return result, err
}

// InsertMany 插入多个文档
func (c *Client) InsertMany(ctx context.Context, collection string, documents []interface{}) (*mongo.InsertManyResult, error) {
	start := time.Now()
	result, err := c.Collection(collection).InsertMany(ctx, documents)
	c.recordMetrics("insert_many", collection, err, start)
	return result, err
}

// FindOne 查询单个文档
func (c *Client) FindOne(ctx context.Context, collection string, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	start := time.Now()
	result := c.Collection(collection).FindOne(ctx, filter, opts...)
	c.recordMetrics("find_one", collection, result.Err(), start)
	return result
}

// Find 查询多个文档
func (c *Client) Find(ctx context.Context, collection string, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	start := time.Now()
	cursor, err := c.Collection(collection).Find(ctx, filter, opts...)
	c.recordMetrics("find", collection, err, start)
	return cursor, err
}

// UpdateOne 更新单个文档
func (c *Client) UpdateOne(ctx context.Context, collection string, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	start := time.Now()
	result, err := c.Collection(collection).UpdateOne(ctx, filter, update, opts...)
	c.recordMetrics("update_one", collection, err, start)
	return result, err
}

// UpdateMany 更新多个文档
func (c *Client) UpdateMany(ctx context.Context, collection string, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	start := time.Now()
	result, err := c.Collection(collection).UpdateMany(ctx, filter, update, opts...)
	c.recordMetrics("update_many", collection, err, start)
	return result, err
}

// DeleteOne 删除单个文档
func (c *Client) DeleteOne(ctx context.Context, collection string, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	start := time.Now()
	result, err := c.Collection(collection).DeleteOne(ctx, filter, opts...)
	c.recordMetrics("delete_one", collection, err, start)
	return result, err
}

// DeleteMany 删除多个文档
func (c *Client) DeleteMany(ctx context.Context, collection string, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	start := time.Now()
	result, err := c.Collection(collection).DeleteMany(ctx, filter, opts...)
	c.recordMetrics("delete_many", collection, err, start)
	return result, err
}

// CountDocuments 统计文档数量
func (c *Client) CountDocuments(ctx context.Context, collection string, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	start := time.Now()
	count, err := c.Collection(collection).CountDocuments(ctx, filter, opts...)
	c.recordMetrics("count_documents", collection, err, start)
	return count, err
}

// Aggregate 聚合查询
func (c *Client) Aggregate(ctx context.Context, collection string, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	start := time.Now()
	cursor, err := c.Collection(collection).Aggregate(ctx, pipeline, opts...)
	c.recordMetrics("aggregate", collection, err, start)
	return cursor, err
}

// recordMetrics 记录指标
func (c *Client) recordMetrics(operation, collection string, err error, start time.Time) {
	if !c.config.EnableMetrics || c.metrics == nil {
		return
	}

	// 记录操作耗时
	duration := time.Since(start).Seconds()
	c.metrics.Histogram(
		"mongodb_operation_duration_seconds",
		"MongoDB operation duration in seconds",
		[]string{"operation", "collection", "service"},
		[]float64{0.1, 0.5, 1, 2, 5, 10},
	).WithLabelValues(operation, collection, c.config.ServiceName).Observe(duration)

	// 记录操作计数
	c.metrics.Counter(
		"mongodb_operations_total",
		"Total number of MongoDB operations",
		[]string{"operation", "collection", "status", "service"},
	).WithLabelValues(operation, collection, getStatus(err), c.config.ServiceName).Inc()
}

// getStatus 获取操作状态
func getStatus(err error) string {
	if err == nil {
		return "success"
	}
	return "error"
}

// 错误定义
var (
	ErrNotFound        = errors.New("document not found")
	ErrInvalidID       = errors.New("invalid id")
	ErrDuplicateKey    = errors.New("duplicate key error")
	ErrInvalidArgument = errors.New("invalid argument")
)

// ConvertID 转换 ID 为 ObjectID
func ConvertID(id string) (primitive.ObjectID, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, ErrInvalidID
	}
	return objectID, nil
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, collection string, keys bson.D, opts ...*options.IndexOptions) (string, error) {
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index(),
	}

	for _, opt := range opts {
		indexModel.Options = opt
	}

	return c.Collection(collection).Indexes().CreateOne(ctx, indexModel)
}

// DropIndex 删除索引
func (c *Client) DropIndex(ctx context.Context, collection string, name string) error {
	_, err := c.Collection(collection).Indexes().DropOne(ctx, name)
	return err
}

// ListIndexes 列出索引
func (c *Client) ListIndexes(ctx context.Context, collection string) (*mongo.Cursor, error) {
	return c.Collection(collection).Indexes().List(ctx)
}
