# Calorie

一个轻量级的 Golang 工具包集合，专注于提供高性能、易用的开发工具。

## 功能特性

- 基于 logrus 的日志工具，支持日志轮转
- Gin 框架的响应中间件
- MongoDB 客户端封装
- Redis 客户端封装
- 统一的错误处理机制

## 设计理念

- 简单易用：提供简洁的 API 接口
- 高性能：底层使用高性能的库实现
- 可扩展：支持自定义配置和扩展
- 稳定性：完善的错误处理和资源管理

## 安装

```bash
go get github.com/NHYCRaymond/calorie
```

## 使用示例

### 错误处理

```go
import "github.com/NHYCRaymond/calorie/pkg/err"

// 创建错误
e := err.New(err.CodeNotFound, "user not found", "user_id: 123")

// 添加错误详情
e = e.WithDetails("username: test")

// 判断错误类型
if err.Is(e, err.CodeNotFound) {
    // 处理未找到错误
}

// 获取错误消息
message := err.GetMessage(err.CodeNotFound)
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
})

// 使用日志
logger.Info("This is an info message")
```

### Gin 响应中间件

```go
import (
    "github.com/NHYCRaymond/calorie/pkg/gin"
    "github.com/NHYCRaymond/calorie/pkg/err"
)

// 在路由中使用
router := gin.Default()
router.Use(gin.ResponseMiddleware())

// 在处理器中使用
func handler(c *gin.Context) {
    // 成功响应
    gin.Success(c, gin.H{
        "data": "your data",
    })

    // 错误响应
    e := err.New(err.CodeNotFound, "user not found")
    gin.Error(c, e.Code, e.Message)
}
```

### MongoDB 客户端

```go
import "github.com/NHYCRaymond/calorie/pkg/mongodb"

// 初始化客户端
client, err := mongodb.NewClient(&mongodb.Config{
    URI:      "mongodb://localhost:27017",
    Database: "test",
})
```

### Redis 客户端

```go
import "github.com/NHYCRaymond/calorie/pkg/redis"

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