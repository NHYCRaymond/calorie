# Calorie

一个轻量级的 Golang 工具包集合，专注于提供高性能、易用的开发工具。

## 功能特性

- 基于 logrus 的日志工具，支持日志轮转
- Gin 框架的响应中间件
- MongoDB 客户端封装
- Redis 客户端封装

## 设计理念

- 简单易用：提供简洁的 API 接口
- 高性能：底层使用高性能的库实现
- 可扩展：支持自定义配置和扩展
- 稳定性：完善的错误处理和资源管理

## 安装

```bash
go get github.com/raymond/calorie
```

## 使用示例

### 日志工具

```go
import "github.com/raymond/calorie/pkg/logger"

// 初始化日志配置
logger.InitLogger(&logger.Config{
    LogPath:     "./logs",
    MaxSize:     64, // MB
    MaxBackups:  7,
    MaxAge:      30, // days
    Compress:    true,
})

// 使用日志
logger.Info("This is an info message")
```

### Gin 响应中间件

```go
import "github.com/raymond/calorie/pkg/gin"

// 在路由中使用
router := gin.Default()
router.Use(gin.ResponseMiddleware())

// 在处理器中使用
func handler(c *gin.Context) {
    gin.Success(c, gin.H{
        "data": "your data",
    })
}
```

### MongoDB 客户端

```go
import "github.com/raymond/calorie/pkg/mongodb"

// 初始化客户端
client, err := mongodb.NewClient(&mongodb.Config{
    URI:      "mongodb://localhost:27017",
    Database: "test",
})
```

### Redis 客户端

```go
import "github.com/raymond/calorie/pkg/redis"

// 初始化客户端
client, err := redis.NewClient(&redis.Config{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})
```

## 贡献指南

欢迎提交 Issue 和 Pull Request 来帮助改进这个项目。

## 许可证

MIT License 