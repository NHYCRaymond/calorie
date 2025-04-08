# 数据库连接示例

这个目录包含了 MongoDB 和 Redis 的连接示例代码。

## MongoDB 示例

MongoDB 示例展示了如何：
1. 创建 MongoDB 客户端
2. 插入文档
3. 查询文档
4. 使用集合

运行示例：
```bash
cd mongodb
go run main.go
```

## Redis 示例

Redis 示例展示了如何：
1. 创建 Redis 客户端
2. 基本的键值操作
3. 使用 Pipeline
4. 使用事务

运行示例：
```bash
cd redis
go run main.go
```

## 注意事项

1. 运行示例前，请确保已安装并启动了相应的数据库服务
2. 默认连接配置：
   - MongoDB: localhost:27017
   - Redis: localhost:6379
3. 如果需要修改连接配置，请编辑相应的 main.go 文件
4. 示例代码包含了错误处理和资源清理
5. 所有操作都使用了上下文超时控制
6. 启用了指标收集功能 