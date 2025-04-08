// Package redis provides a Redis client wrapper with connection pooling and common operations.
// It offers a simple interface for Redis key-value operations.
package redis

import (
	"context"
	"strings"
	"time"

	"github.com/NHYCRaymond/calorie/pkg/errors"
	"github.com/NHYCRaymond/calorie/pkg/metrics"
	"github.com/go-redis/redis/v8"
)

// Config Redis 配置
type Config struct {
	Addr            string
	Password        string
	DB              int
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolSize        int
	MinIdleConns    int
	MaxConnAge      time.Duration
	// 是否启用指标收集
	EnableMetrics bool
	// 服务名称
	ServiceName string
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	MaxRetries:      3,
	MinRetryBackoff: 8 * time.Millisecond,
	MaxRetryBackoff: 512 * time.Millisecond,
	DialTimeout:     5 * time.Second,
	ReadTimeout:     3 * time.Second,
	WriteTimeout:    3 * time.Second,
	PoolSize:        10,
	MinIdleConns:    5,
	MaxConnAge:      30 * time.Minute,
	EnableMetrics:   true,
	ServiceName:     "redis",
}

// Client Redis 客户端
type Client struct {
	client  *redis.Client
	config  *Config
	metrics *metrics.Client
}

// NewClient 创建新的 Redis 客户端
func NewClient(config *Config, metricsClient *metrics.Client) (*Client, error) {
	if config == nil {
		config = DefaultConfig
	}

	client := redis.NewClient(&redis.Options{
		Addr:            config.Addr,
		Password:        config.Password,
		DB:              config.DB,
		MaxRetries:      config.MaxRetries,
		MinRetryBackoff: config.MinRetryBackoff,
		MaxRetryBackoff: config.MaxRetryBackoff,
		DialTimeout:     config.DialTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
		MaxConnAge:      config.MaxConnAge,
	})

	// 检查连接
	ctx, cancel := context.WithTimeout(context.Background(), config.DialTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	// 如果启用了指标收集但没有传入 metrics client，则禁用指标收集
	if config.EnableMetrics && metricsClient == nil {
		config.EnableMetrics = false
	}

	return &Client{
		client:  client,
		config:  config,
		metrics: metricsClient,
	}, nil
}

// 错误定义
var (
	// ErrKeyNotFound 键不存在
	ErrKeyNotFound = errors.New(errors.CodeNotFound, "key not found")
	// ErrInvalidType 类型错误
	ErrInvalidType = errors.New(errors.CodeError, "invalid type")
	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errors.New(errors.CodeServerError, "connection failed")
	// ErrOperationTimeout 操作超时
	ErrOperationTimeout = errors.New(errors.CodeServerError, "operation timeout")
	// ErrInvalidArgument 参数错误
	ErrInvalidArgument = errors.New(errors.CodeError, "invalid argument")
	// ErrRedisError Redis 错误
	ErrRedisError = errors.New(errors.CodeServerError, "redis error")
)

// wrapError 包装错误
func wrapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Redis 原生错误
	if err == redis.Nil {
		return errors.New(errors.CodeNotFound, "key not found").WithDetails(operation)
	}

	// 网络错误
	if strings.Contains(err.Error(), "connection refused") {
		return errors.New(errors.CodeServerError, "connection refused").WithDetails(operation)
	}

	// 超时错误
	if strings.Contains(err.Error(), "timeout") {
		return errors.New(errors.CodeServerError, "operation timeout").WithDetails(operation)
	}

	// 其他错误
	return errors.New(errors.CodeServerError, err.Error()).WithDetails(operation)
}

// operation 定义 Redis 操作
type operation struct {
	name      string
	client    *redis.Client
	metrics   *metrics.Client
	config    *Config
	startTime time.Time
}

// newOperation 创建新的操作
func (c *Client) newOperation(name string) *operation {
	return &operation{
		name:      name,
		client:    c.client,
		metrics:   c.metrics,
		config:    c.config,
		startTime: time.Now(),
	}
}

// end 结束操作并记录指标
func (op *operation) end(err error) {
	if !op.config.EnableMetrics || op.metrics == nil {
		return
	}

	duration := time.Since(op.startTime).Seconds()
	op.metrics.Histogram(
		"redis_operation_duration_seconds",
		"Redis operation duration in seconds",
		[]string{"operation", "service"},
		[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
	).WithLabelValues(op.name, op.config.ServiceName).Observe(duration)

	op.metrics.Counter(
		"redis_operations_total",
		"Total number of Redis operations",
		[]string{"operation", "status", "service"},
	).WithLabelValues(op.name, getStatus(err), op.config.ServiceName).Inc()
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	op := c.newOperation("get")
	result, err := c.client.Get(ctx, key).Result()
	op.end(err)
	if err != nil {
		return "", wrapError(err, "get")
	}
	return result, nil
}

// Set 设置值
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	op := c.newOperation("set")
	err := c.client.Set(ctx, key, value, expiration).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "set")
	}
	return nil
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	op := c.newOperation("del")
	err := c.client.Del(ctx, keys...).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "del")
	}
	return nil
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	op := c.newOperation("exists")
	result, err := c.client.Exists(ctx, keys...).Result()
	op.end(err)
	if err != nil {
		return 0, wrapError(err, "exists")
	}
	return result, nil
}

// Expire 设置过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	op := c.newOperation("expire")
	result, err := c.client.Expire(ctx, key, expiration).Result()
	op.end(err)
	if err != nil {
		return false, wrapError(err, "expire")
	}
	return result, nil
}

// TTL 获取过期时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	op := c.newOperation("ttl")
	result, err := c.client.TTL(ctx, key).Result()
	op.end(err)
	if err != nil {
		return 0, wrapError(err, "ttl")
	}
	return result, nil
}

// Incr 自增
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	op := c.newOperation("incr")
	result, err := c.client.Incr(ctx, key).Result()
	op.end(err)
	if err != nil {
		return 0, wrapError(err, "incr")
	}
	return result, nil
}

// Decr 自减
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	op := c.newOperation("decr")
	result, err := c.client.Decr(ctx, key).Result()
	op.end(err)
	if err != nil {
		return 0, wrapError(err, "decr")
	}
	return result, nil
}

// HGet 获取哈希字段值
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	op := c.newOperation("hget")
	result, err := c.client.HGet(ctx, key, field).Result()
	op.end(err)
	if err != nil {
		return "", wrapError(err, "hget")
	}
	return result, nil
}

// HSet 设置哈希字段值
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	op := c.newOperation("hset")
	err := c.client.HSet(ctx, key, values...).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "hset")
	}
	return nil
}

// HDel 删除哈希字段
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	op := c.newOperation("hdel")
	err := c.client.HDel(ctx, key, fields...).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "hdel")
	}
	return nil
}

// HGetAll 获取所有哈希字段
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	op := c.newOperation("hgetall")
	result, err := c.client.HGetAll(ctx, key).Result()
	op.end(err)
	if err != nil {
		return nil, wrapError(err, "hgetall")
	}
	return result, nil
}

// LPush 列表头部插入
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	op := c.newOperation("lpush")
	err := c.client.LPush(ctx, key, values...).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "lpush")
	}
	return nil
}

// RPush 列表尾部插入
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	op := c.newOperation("rpush")
	err := c.client.RPush(ctx, key, values...).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "rpush")
	}
	return nil
}

// LPop 列表头部弹出
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	op := c.newOperation("lpop")
	result, err := c.client.LPop(ctx, key).Result()
	op.end(err)
	if err != nil {
		return "", wrapError(err, "lpop")
	}
	return result, nil
}

// RPop 列表尾部弹出
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	op := c.newOperation("rpop")
	result, err := c.client.RPop(ctx, key).Result()
	op.end(err)
	if err != nil {
		return "", wrapError(err, "rpop")
	}
	return result, nil
}

// LLen 获取列表长度
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	op := c.newOperation("llen")
	result, err := c.client.LLen(ctx, key).Result()
	op.end(err)
	if err != nil {
		return 0, wrapError(err, "llen")
	}
	return result, nil
}

// SAdd 集合添加成员
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	op := c.newOperation("sadd")
	err := c.client.SAdd(ctx, key, members...).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "sadd")
	}
	return nil
}

// SRem 集合移除成员
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) error {
	op := c.newOperation("srem")
	err := c.client.SRem(ctx, key, members...).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "srem")
	}
	return nil
}

// SMembers 获取集合所有成员
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	op := c.newOperation("smembers")
	result, err := c.client.SMembers(ctx, key).Result()
	op.end(err)
	if err != nil {
		return nil, wrapError(err, "smembers")
	}
	return result, nil
}

// SIsMember 判断成员是否在集合中
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	op := c.newOperation("sismember")
	result, err := c.client.SIsMember(ctx, key, member).Result()
	op.end(err)
	if err != nil {
		return false, wrapError(err, "sismember")
	}
	return result, nil
}

// ZAdd 有序集合添加成员
func (c *Client) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	op := c.newOperation("zadd")
	err := c.client.ZAdd(ctx, key, members...).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "zadd")
	}
	return nil
}

// ZRem 有序集合移除成员
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) error {
	op := c.newOperation("zrem")
	err := c.client.ZRem(ctx, key, members...).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "zrem")
	}
	return nil
}

// ZRange 获取有序集合成员
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	op := c.newOperation("zrange")
	result, err := c.client.ZRange(ctx, key, start, stop).Result()
	op.end(err)
	if err != nil {
		return nil, wrapError(err, "zrange")
	}
	return result, nil
}

// ZRangeWithScores 获取有序集合成员及分数
func (c *Client) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	op := c.newOperation("zrange_with_scores")
	result, err := c.client.ZRangeWithScores(ctx, key, start, stop).Result()
	op.end(err)
	if err != nil {
		return nil, wrapError(err, "zrange_with_scores")
	}
	return result, nil
}

// Publish 发布消息
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	op := c.newOperation("publish")
	err := c.client.Publish(ctx, channel, message).Err()
	op.end(err)
	if err != nil {
		return wrapError(err, "publish")
	}
	return nil
}

// Subscribe 订阅频道
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.client.Subscribe(ctx, channels...)
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.client.Close()
}

// getStatus 获取操作状态
func getStatus(err error) string {
	if err == nil {
		return "success"
	}
	return "error"
}
