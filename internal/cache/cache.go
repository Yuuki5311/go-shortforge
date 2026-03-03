package cache

import "context"

// 缓存接口：定义基础的读写与删除能力
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, val string, ttlSeconds int) error
	Del(ctx context.Context, key string) error
}
