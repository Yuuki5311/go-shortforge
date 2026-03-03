package cache

import (
	"context"
	"time"

	"github.com/ego-component/eredis"
)

// Redis 缓存实现：封装基础的 Get/Set/Del 操作
type RedisCache struct {
	client *eredis.Component
}

// 创建缓存实例
func NewRedisCache(c *eredis.Component) *RedisCache {
	return &RedisCache{client: c}
}

// 读取缓存
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key)
}

// 写入缓存（附带过期时间）
func (r *RedisCache) Set(ctx context.Context, key, val string, ttlSeconds int) error {
	return r.client.Set(ctx, key, val, time.Duration(ttlSeconds)*time.Second)
}

// 删除缓存键
func (r *RedisCache) Del(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key)
	return err
}
