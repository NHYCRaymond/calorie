// Package mysql provides a MySQL client wrapper with connection pooling and authentication support.
// It simplifies MySQL operations with a clean and consistent API.
//
// 协程安全说明：
// 1. Client 实例是协程安全的，可以在多个 goroutine 中共享
// 2. 连接池管理是线程安全的
// 3. 事务操作是隔离的，每个事务都有自己的上下文
// 4. 查询和更新操作是原子的
package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/NHYCRaymond/calorie/pkg/metrics"
	_ "github.com/go-sql-driver/mysql"
)

// Config MySQL 配置
type Config struct {
	URI             string
	Database        string
	Username        string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	// 是否启用指标收集
	EnableMetrics bool
	// 服务名称
	ServiceName string
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	MaxOpenConns:    100,
	MaxIdleConns:    10,
	ConnMaxLifetime: 30 * time.Minute,
	ConnMaxIdleTime: 10 * time.Minute,
	EnableMetrics:   true,
	ServiceName:     "mysql",
}

// Client MySQL 客户端
type Client struct {
	db      *sql.DB
	config  *Config
	metrics *metrics.Client
}

// NewClient 创建新的 MySQL 客户端
// 注意：返回的 Client 实例是协程安全的，可以在多个 goroutine 中共享
func NewClient(config *Config, metricsClient *metrics.Client) (*Client, error) {
	if config == nil {
		config = DefaultConfig
	}

	// 构建 DSN
	dsn := config.Username + ":" + config.Password + "@" + config.URI + "/" + config.Database

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 设置连接池
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 检查连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &Client{
		db:      db,
		config:  config,
		metrics: metricsClient,
	}, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.db.Close()
}

// WithTransaction 执行事务
// 注意：事务操作是隔离的，每个事务都有自己的上下文
// 建议：不要在事务中执行长时间运行的操作，以免阻塞其他事务
func (c *Client) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.New(err.Error() + ": " + rbErr.Error())
		}
		return err
	}

	return tx.Commit()
}

// Query 执行查询
// 注意：查询操作是原子的，但返回的 Rows 对象不是协程安全的
// 建议：每个 goroutine 使用自己的 Rows 对象
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := c.db.QueryContext(ctx, query, args...)
	c.recordMetrics("query", err, start)
	return rows, err
}

// QueryRow 执行单行查询
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := c.db.QueryRowContext(ctx, query, args...)
	c.recordMetrics("query_row", row.Err(), start)
	return row
}

// Exec 执行非查询操作
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := c.db.ExecContext(ctx, query, args...)
	c.recordMetrics("exec", err, start)
	return result, err
}

// Prepare 准备语句
func (c *Client) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	start := time.Now()
	stmt, err := c.db.PrepareContext(ctx, query)
	c.recordMetrics("prepare", err, start)
	return stmt, err
}

// Begin 开始事务
func (c *Client) Begin(ctx context.Context) (*sql.Tx, error) {
	start := time.Now()
	tx, err := c.db.BeginTx(ctx, nil)
	c.recordMetrics("begin", err, start)
	return tx, err
}

// recordMetrics 记录指标
func (c *Client) recordMetrics(operation string, err error, start time.Time) {
	if !c.config.EnableMetrics || c.metrics == nil {
		return
	}

	// 记录操作耗时
	duration := time.Since(start).Seconds()
	c.metrics.Histogram(
		"mysql_operation_duration_seconds",
		"MySQL operation duration in seconds",
		[]string{"operation", "service"},
		[]float64{0.1, 0.5, 1, 2, 5, 10},
	).WithLabelValues(operation, c.config.ServiceName).Observe(duration)

	// 记录操作计数
	c.metrics.Counter(
		"mysql_operations_total",
		"Total number of MySQL operations",
		[]string{"operation", "status", "service"},
	).WithLabelValues(operation, getStatus(err), c.config.ServiceName).Inc()
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
	ErrNoRows          = sql.ErrNoRows
	ErrTxDone          = sql.ErrTxDone
	ErrConnDone        = sql.ErrConnDone
	ErrInvalidArgument = errors.New("invalid argument")
)
