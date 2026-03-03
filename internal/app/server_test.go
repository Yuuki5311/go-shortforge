package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"shorturl/internal/model"
	"shorturl/internal/repo"
	"shorturl/internal/service"
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

type fakeCache struct{ m map[string]string }

func (f *fakeCache) Get(ctx context.Context, key string) (string, error) { return f.m[key], nil }
func (f *fakeCache) Set(ctx context.Context, key, val string, ttl int) error {
	if f.m == nil {
		f.m = map[string]string{}
	}
	f.m[key] = val
	return nil
}
func (f *fakeCache) Del(ctx context.Context, key string) error {
	delete(f.m, key)
	return nil
}

func TestRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newFakeRepo()
	ca := &fakeCache{m: map[string]string{}}
	s := service.New(r, ca, 60)
	code, err := s.Shorten(context.Background(), "https://example.com/redirect")
	if err != nil {
		t.Fatalf("shorten err: %v", err)
	}
	setService(s)
	engine := gin.New()
	engine.GET("/u/:code", redirect)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/u/"+code, nil)
	engine.ServeHTTP(w, req)
	if w.Code != 302 {
		t.Fatalf("expect 302, got %d body=%s", w.Code, w.Body.String())
	}
	loc := w.Header().Get("Location")
	if loc != "https://example.com/redirect" {
		t.Fatalf("expect redirect to long url, got %s", loc)
	}
}

