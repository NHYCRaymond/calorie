# Calorie

一个轻量级的 Golang 工具包集合，专注于提供高性能、易用的开发工具。

## 功能特性

- 基于 logrus 的日志工具，支持日志轮转和多级日志
- Gin 框架的中间件集合，包括请求追踪、指标收集等
- MongoDB 客户端封装，支持连接池和常用操作
- MySQL 客户端封装，支持连接池和事务管理
- Redis 客户端封装，支持连接池和完整的数据类型操作
- Prometheus 指标收集，支持自定义指标和多种指标类型
- 统一的错误处理机制，支持错误码和错误详情

## 设计理念

- 简单易用：提供简洁的 API 接口
- 高性能：底层使用高性能的库实现
- 可扩展：支持自定义配置和扩展
- 稳定性：完善的错误处理和资源管理
- 可观测性：内置监控和追踪支持

## 安装

```bash
go get github.com/NHYCRaymond/calorie
```

## 使用示例

### 错误处理

```go
import "github.com/NHYCRaymond/calorie/pkg/errors"

// 创建错误
e := errors.New(errors.CodeNotFound, "user not found")

// 添加错误详情
e = e.WithDetails("user_id: 123", "username: test")

// 判断错误类型
if errors.Is(e, errors.CodeNotFound) {
    // 处理未找到错误
}

// 获取错误消息
message := errors.GetMessage(errors.CodeNotFound)
```

### 日志工具

```go
import "github.com/NHYCRaymond/calorie/pkg/logger"

// 初始化日志配置
logger.InitLogger(&logger.Config{
    LogPath:     "./logs",
    MaxSize:     64, // MB
    MaxBackups:  7,
    MaxAge:      30, // days
    Compress:    true,
    Level:       "info",
    ServiceName: "user-service",
})

// 使用日志
logger.WithFields(logger.Fields{
    "user_id": "123",
    "action": "login",
}).Info("用户登录成功")

// 错误日志
logger.WithError(err).Error("操作失败")
```

### Gin 中间件

```go
import (
    "github.com/NHYCRaymond/calorie/pkg/gin"
    "github.com/NHYCRaymond/calorie/pkg/errors"
)

// 创建路由
router := gin.Default()

// 使用中间件
router.Use(
    gin.RequestTracing(),   // 请求追踪
    gin.RequestMetrics(),   // 请求指标
    gin.RequestTimeout(5 * time.Second), // 请求超时
    gin.Recovery(),         // 错误恢复
)

// 在处理器中使用
func handler(c *gin.Context) {
    // 成功响应
    gin.Success(c, gin.H{
        "data": "your data",
    })

    // 错误响应
    e := errors.New(errors.CodeNotFound, "user not found")
    gin.Error(c, e)
}
```

### MongoDB 客户端

```go
import "github.com/NHYCRaymond/calorie/pkg/mongodb"

// 初始化客户端
client, err := mongodb.NewClient(&mongodb.Config{
    URI:             "mongodb://localhost:27017",
    Database:        "test",
    MaxPoolSize:     100,
    MinPoolSize:     10,
    MaxConnIdleTime: 30 * time.Minute,
    EnableMetrics:   true,
    ServiceName:     "user-service",
})

// 使用客户端
ctx := context.Background()
collection := client.Collection("users")

// 插入文档
doc := bson.D{{"name", "test"}, {"age", 18}}
result, err := collection.InsertOne(ctx, doc)

// 查询文档
var user User
err = collection.FindOne(ctx, bson.D{{"name", "test"}}).Decode(&user)
```

### MySQL 客户端

```go
import "github.com/NHYCRaymond/calorie/pkg/mysql"

// 初始化客户端
client, err := mysql.NewClient(&mysql.Config{
    URI:             "tcp(localhost:3306)",
    Database:        "test",
    Username:        "root",
    Password:        "password",
    MaxOpenConns:    100,
    MaxIdleConns:    10,
    ConnMaxLifetime: 30 * time.Minute,
    ConnMaxIdleTime: 10 * time.Minute,
    EnableMetrics:   true,
    ServiceName:     "user-service",
})

// 使用客户端
ctx := context.Background()

// 查询操作
rows, err := client.Query(ctx, "SELECT * FROM users WHERE age > ?", 18)
if err != nil {
    return err
}
defer rows.Close()

// 单行查询
var user User
err = client.QueryRow(ctx, "SELECT * FROM users WHERE id = ?", 1).Scan(&user.ID, &user.Name, &user.Age)

// 执行操作
result, err := client.Exec(ctx, "INSERT INTO users (name, age) VALUES (?, ?)", "test", 18)

// 事务操作
err = client.WithTransaction(ctx, func(tx *sql.Tx) error {
    // 在事务中执行操作
    _, err := tx.ExecContext(ctx, "UPDATE users SET age = ? WHERE id = ?", 19, 1)
    if err != nil {
        return err
    }
    return nil
})
```

### Redis 客户端

```go
import "github.com/NHYCRaymond/calorie/pkg/redis"

// 初始化客户端
client, err := redis.NewClient(&redis.Config{
    Addr:            "localhost:6379",
    Password:        "",
    DB:             0,
    MaxRetries:      3,
    PoolSize:        10,
    MinIdleConns:    5,
    EnableMetrics:   true,
    ServiceName:     "user-service",
})

// 字符串操作
ctx := context.Background()
err = client.Set(ctx, "key", "value", time.Hour)
value, err := client.Get(ctx, "key")

// 哈希操作
err = client.HSet(ctx, "hash", "field", "value")
value, err = client.HGet(ctx, "hash", "field")

// 列表操作
err = client.LPush(ctx, "list", "value1", "value2")
value, err = client.LPop(ctx, "list")

// 集合操作
err = client.SAdd(ctx, "set", "member1", "member2")
exists, err := client.SIsMember(ctx, "set", "member1")

// 有序集合操作
err = client.ZAdd(ctx, "zset", &redis.Z{Score: 1, Member: "member1"})
members, err := client.ZRange(ctx, "zset", 0, -1)
```

### 监控指标

```go
import "github.com/NHYCRaymond/calorie/pkg/metrics"

// 创建监控客户端
client := metrics.NewClient(&metrics.Config{
    Enabled:     true,
    Addr:        ":9090",
    Path:        "/metrics",
    ServiceName: "user-service",
})

// 计数器
counter := client.Counter(
    "http_requests_total",
    "Total number of HTTP requests",
    []string{"method", "path", "status"},
)
counter.WithLabelValues("GET", "/api/users", "200").Inc()

// 直方图
histogram := client.Histogram(
    "http_request_duration_seconds",
    "HTTP request duration in seconds",
    []string{"method", "path"},
    []float64{0.1, 0.5, 1, 2, 5},
)
histogram.WithLabelValues("GET", "/api/users").Observe(0.2)

// 仪表盘
gauge := client.Gauge(
    "memory_usage_bytes",
    "Memory usage in bytes",
    []string{"service"},
)
gauge.WithLabelValues("user-service").Set(1024 * 1024 * 100)

// 摘要
summary := client.Summary(
    "request_size_bytes",
    "Request size in bytes",
    []string{"method", "path"},
    map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
)
summary.WithLabelValues("POST", "/api/users").Observe(1024)
```

## 包说明

### errors
统一的错误处理包，提供错误码定义、错误创建、错误详情添加等功能。

### logger
基于 logrus 的日志工具，支持日志轮转、多级日志、结构化日志等功能。

### gin
Gin 框架的中间件集合，包括请求追踪、指标收集、超时控制、错误处理等。

### mongodb
MongoDB 客户端封装，提供连接池管理、常用操作封装、指标收集等功能。

### mysql
MySQL 客户端封装，提供连接池管理、事务支持、指标收集等功能。

### redis
Redis 客户端封装，支持所有数据类型操作、连接池管理、指标收集等功能。

### metrics
Prometheus 指标收集工具，支持计数器、仪表盘、直方图、摘要等多种指标类型。

## 贡献指南

欢迎提交 Issue 和 Pull Request 来帮助改进这个项目。在提交代码前，请确保：

1. 代码符合 Go 编码规范
2. 添加了必要的单元测试
3. 更新了相关文档
4. 通过了所有测试

## 许可证

MIT License 