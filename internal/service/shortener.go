package service

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/rand"
	"time"

	"shorturl/internal/cache"
	"shorturl/internal/repo"
)

// 业务服务：封装短链生成、解析、删除与批量处理逻辑
type Service struct {
	repo        repo.Repository // 数据访问（MySQL）
	cache       cache.Cache     // 缓存访问（Redis）
	cacheTTLsec int             // 缓存过期时间（秒）
}

// 创建服务实例
func New(repo repo.Repository, cache cache.Cache, cacheTTLsec int) *Service {
	return &Service{repo: repo, cache: cache, cacheTTLsec: cacheTTLsec}
}

// 生成短链（幂等）：
// - 先按长链查询，存在则直接返回已有短码并写缓存
// - 不存在则生成 7 位短码，最多尝试 5 次避免冲突；入库并写缓存
func (s *Service) Shorten(ctx context.Context, longURL string) (string, error) {
	// idempotent: if exists, return existing code
	if exist, err := s.repo.FindByLong(ctx, longURL); err == nil && exist != nil {
		_ = s.cache.Set(ctx, s.cacheKey(exist.Code), exist.LongURL, s.cacheTTLsec)
		return exist.Code, nil
	}

	var code string
	for i := 0; i < 5; i++ {
		code = genCode(longURL, i)
		u, err := s.repo.FindByCode(ctx, code)
		if err != nil {
			return "", err
		}
		if u == nil {
			break
		}
	}
	if code == "" {
		return "", errors.New("generate code failed")
	}
	item, err := s.repo.Create(ctx, code, longURL)
	if err != nil {
		return "", err
	}
	_ = s.cache.Set(ctx, s.cacheKey(item.Code), item.LongURL, s.cacheTTLsec)
	return item.Code, nil
}

// 解析短码：
// - 优先读缓存，未命中则查库并回填缓存
func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	if s.cache != nil {
		if v, err := s.cache.Get(ctx, s.cacheKey(code)); err == nil && v != "" {
			return v, nil
		}
	}
	item, err := s.repo.FindByCode(ctx, code)
	if err != nil || item == nil {
		return "", err
	}
	_ = s.cache.Set(ctx, s.cacheKey(code), item.LongURL, s.cacheTTLsec)
	return item.LongURL, nil
}

// 删除短码：先删库，再删缓存键
func (s *Service) Delete(ctx context.Context, code string) error {
	if err := s.repo.DeleteByCode(ctx, code); err != nil {
		return err
	}
	_ = s.cache.Del(ctx, s.cacheKey(code))
	return nil
}

// 批量生成：
// - 输入去重；已存在复用短码
// - 为未存在生成候选码并避免冲突；批量入库与回填缓存
func (s *Service) BatchShorten(ctx context.Context, urls []string) (map[string]string, error) {
	// deduplicate input
	seen := map[string]struct{}{}
	urls2 := make([]string, 0, len(urls))
	for _, u := range urls {
		if _, ok := seen[u]; !ok {
			seen[u] = struct{}{}
			urls2 = append(urls2, u)
		}
	}
	// pre-check existing
	result := map[string]string{}
	toCreate := map[string]string{}
	for _, u := range urls2 {
		if exist, err := s.repo.FindByLong(ctx, u); err == nil && exist != nil {
			result[u] = exist.Code
		} else {
			// generate tentative unique code
			for i := 0; i < 5; i++ {
				code := genCode(u, i)
				if exist2, _ := s.repo.FindByCode(ctx, code); exist2 == nil {
					toCreate[code] = u
					result[u] = code
					break
				}
			}
		}
	}
	if len(toCreate) > 0 {
		list, err := s.repo.BatchCreate(ctx, toCreate)
		if err != nil {
			return nil, err
		}
		for _, it := range list {
			_ = s.cache.Set(ctx, s.cacheKey(it.Code), it.LongURL, s.cacheTTLsec)
		}
	}
	return result, nil
}

// 生成缓存键
func (s *Service) cacheKey(code string) string {
	return "shorturl:code:" + code
}

// 短码生成算法：
// - 基于长链的 SHA256 + 当前时间 + salt 构造随机种子
// - 使用 base62 字符集生成固定长度（7）的短码
func genCode(longURL string, salt int) string {
	h := sha256.Sum256([]byte(longURL))
	seed := int64(binary.BigEndian.Uint64(h[:8])) + int64(salt) + time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	n := 7
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(b)
}
