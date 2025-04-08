package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NHYCRaymond/calorie/pkg/logger"
	"github.com/NHYCRaymond/calorie/pkg/metrics"
	"github.com/NHYCRaymond/calorie/pkg/redis"
)

func main() {
	// 初始化日志
	logger.InitLogger(&logger.Config{
		LogPath:    "../logs",
		MaxSize:    64, // MB
		MaxBackups: 10,
		MaxAge:     30, // days
		Compress:   true,
	})

	// 创建 metrics 客户端
	metricsClient := metrics.NewClient(&metrics.Config{
		Enabled: true,
		Addr:    ":8080",
		Path:    "/metrics",
	})

	// 创建 Redis 配置
	config := &redis.Config{
		Addr:            "localhost:6379",
		Password:        "",
		DB:              0,
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
		ServiceName:     "redis-example",
	}

	// 创建 Redis 客户端
	client, err := redis.NewClient(config, metricsClient)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 设置键值
	err = client.Set(ctx, "key", "value", 10*time.Minute)
	if err != nil {
		log.Fatalf("Failed to set key: %v", err)
	}

	// 获取键值
	value, err := client.Get(ctx, "key")
	if err != nil {
		log.Fatalf("Failed to get key: %v", err)
	}
	fmt.Printf("Got value: %s\n", value)

	// 使用 Pipeline
	pipe := client.Pipeline()
	pipe.Set(ctx, "key1", "value1", 10*time.Minute)
	pipe.Set(ctx, "key2", "value2", 10*time.Minute)
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Fatalf("Failed to execute pipeline: %v", err)
	}

	// 使用事务
	tx := client.TxPipeline()
	tx.Set(ctx, "key3", "value3", 10*time.Minute)
	tx.Set(ctx, "key4", "value4", 10*time.Minute)
	_, err = tx.Exec(ctx)
	if err != nil {
		log.Fatalf("Failed to execute transaction: %v", err)
	}
}
