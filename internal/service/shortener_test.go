package service

import (
	"context"
	"testing"

	"shorturl/internal/model"
	"shorturl/internal/repo"
)

type fakeRepo struct {
	code2url map[string]string
	url2code map[string]string
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{code2url: map[string]string{}, url2code: map[string]string{}}
}

func (f *fakeRepo) Create(ctx context.Context, code, longURL string) (*model.ShortURL, error) {
	f.code2url[code] = longURL
	f.url2code[longURL] = code
	return &model.ShortURL{Code: code, LongURL: longURL}, nil
}
func (f *fakeRepo) FindByCode(ctx context.Context, code string) (*model.ShortURL, error) {
	if u, ok := f.code2url[code]; ok {
		return &model.ShortURL{Code: code, LongURL: u}, nil
	}
	return nil, nil
}
func (f *fakeRepo) FindByLong(ctx context.Context, longURL string) (*model.ShortURL, error) {
	if c, ok := f.url2code[longURL]; ok {
		return &model.ShortURL{Code: c, LongURL: longURL}, nil
	}
	return nil, nil
}
func (f *fakeRepo) DeleteByCode(ctx context.Context, code string) error {
	if u, ok := f.code2url[code]; ok {
		delete(f.code2url, code)
		delete(f.url2code, u)
	}
	return nil
}
func (f *fakeRepo) BatchCreate(ctx context.Context, items map[string]string) ([]model.ShortURL, error) {
	out := make([]model.ShortURL, 0, len(items))
	for c, u := range items {
		f.code2url[c] = u
		f.url2code[u] = c
		out = append(out, model.ShortURL{Code: c, LongURL: u})
	}
	return out, nil
}

var _ repo.Repository = (*fakeRepo)(nil)

type fakeCache struct {
	m map[string]string
}

func newFakeCache() *fakeCache { return &fakeCache{m: map[string]string{}} }

func (f *fakeCache) Get(ctx context.Context, k string) (string, error) { return f.m[k], nil }
func (f *fakeCache) Set(ctx context.Context, k, v string, ttl int) error {
	f.m[k] = v
	return nil
}
func (f *fakeCache) Del(ctx context.Context, k string) error {
	delete(f.m, k)
	return nil
}

func TestShortenResolveDelete(t *testing.T) {
	r := newFakeRepo()
	c := newFakeCache()
	s := &Service{
		repo:        r,
		cache:       c,
		cacheTTLsec: 60,
	}
	ctx := context.Background()
	code, err := s.Shorten(ctx, "https://example.com/a")
	if err != nil || code == "" {
		t.Fatalf("shorten error: %v", err)
	}
	long, err := s.Resolve(ctx, code)
	if err != nil || long != "https://example.com/a" {
		t.Fatalf("resolve failed, got %s err %v", long, err)
	}
	if err := s.Delete(ctx, code); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if long, _ := s.Resolve(ctx, code); long != "" {
		t.Fatalf("expect empty after delete")
	}
}

func TestBatchShorten(t *testing.T) {
	r := newFakeRepo()
	c := newFakeCache()
	s := &Service{
		repo:        r,
		cache:       c,
		cacheTTLsec: 60,
	}
	ctx := context.Background()
	in := []string{"https://a.com", "https://b.com", "https://a.com"}
	res, err := s.BatchShorten(ctx, in)
	if err != nil {
		t.Fatalf("batch shorten error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expect 2 results, got %d", len(res))
	}
}
