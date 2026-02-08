package oss

import (
	"fmt"
	"sync"
)

var (
	// globalClient 全局 OSS 客户端实例
	globalClient Client
	// once 确保只初始化一次
	once sync.Once
	// initErr 初始化错误
	initErr error
)

// Init 初始化 OSS 客户端
func Init(config *Config) error {
	once.Do(func() {
		globalClient, initErr = NewClient(config)
	})

	return initErr
}

// GetClient 获取全局 OSS 客户端实例
func GetClient() (Client, error) {
	if globalClient == nil {
		return nil, fmt.Errorf("oss client not initialized, please call Init() first")
	}
	return globalClient, nil
}

// MustGetClient 获取全局 OSS 客户端实例（如果未初始化则 panic）
func MustGetClient() Client {
	client, err := GetClient()
	if err != nil {
		panic(err)
	}
	return client
}
