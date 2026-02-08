package cache

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var testRedisClient *redis.Client

// InitTestRedis 初始化测试 Redis 客户端
func InitTestRedis() *redis.Client {
	if testRedisClient != nil {
		return testRedisClient
	}

	// 使用 miniredis 或连接到本地 Redis 进行测试
	// 这里使用本地 Redis 的测试数据库（DB 15）
	testRedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       15, // 使用独立的测试数据库
	})

	// 测试连接
	ctx := context.Background()
	if err := testRedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to test Redis: %v", err)
		log.Println("Using mock Redis client for testing")
		// 如果连接失败，可以使用 miniredis 或 mock
		// 这里简单返回 nil，测试会跳过需要 Redis 的部分
		return nil
	}

	log.Println("Test Redis connected successfully")
	return testRedisClient
}

// CleanupTestRedis 清理测试 Redis 客户端
func CleanupTestRedis() {
	if testRedisClient != nil {
		testRedisClient.Close()
		testRedisClient = nil
	}
}

// CleanupTestRedisData 清理测试 Redis 数据
func CleanupTestRedisData() {
	if testRedisClient == nil {
		return
	}

	ctx := context.Background()
	
	// 清空当前数据库的所有数据
	if err := testRedisClient.FlushDB(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to flush test Redis DB: %v", err)
	}
}
