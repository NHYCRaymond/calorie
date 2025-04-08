package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NHYCRaymond/calorie/pkg/logger"
	"github.com/NHYCRaymond/calorie/pkg/metrics"
	"github.com/NHYCRaymond/calorie/pkg/mongodb"
	"github.com/sirupsen/logrus"
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

	// 添加测试日志
	logger.Info("Starting MongoDB example")
	logger.WithFields(logrus.Fields{
		"service": "mongodb-example",
		"version": "1.0.0",
	}).Info("Initializing MongoDB client")

	// 创建 metrics 客户端
	metricsClient := metrics.NewClient(&metrics.Config{
		Enabled: true,
		Addr:    ":8080",
		Path:    "/metrics",
	})

	// 创建 MongoDB 配置
	config := &mongodb.Config{
		URI:            "mongodb://localhost:27017",
		Database:       "test",
		Username:       "",
		Password:       "",
		MaxPoolSize:    100,
		MinPoolSize:    10,
		ConnectTimeout: 10 * time.Second,
		SocketTimeout:  30 * time.Second,
		EnableMetrics:  true,
		ServiceName:    "mongodb-example",
	}

	// 创建 MongoDB 客户端
	client, err := mongodb.NewClient(config, metricsClient)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}
	defer client.Close()

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 插入文档
	doc := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}
	result, err := client.InsertOne(ctx, "users", doc)
	if err != nil {
		log.Fatalf("Failed to insert document: %v", err)
	}
	fmt.Printf("Inserted document with ID: %v\n", result.InsertedID)

	// 查询文档
	filter := map[string]interface{}{
		"name": "John Doe",
	}
	cursor, err := client.Find(ctx, "users", filter)
	if err != nil {
		log.Fatalf("Failed to find documents: %v", err)
	}
	defer cursor.Close(ctx)

	// 遍历结果
	for cursor.Next(ctx) {
		var result map[string]interface{}
		if err := cursor.Decode(&result); err != nil {
			log.Fatalf("Failed to decode document: %v", err)
		}
		fmt.Printf("Found document: %v\n", result)
	}

	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error: %v", err)
	}
}
