package handler

import (
	"bytes"
	"encoding/json"
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

func TestCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newFakeRepo()
	ca := &fakeCache{m: map[string]string{}}
	s := service.New(r, ca, 60)
	h := New(s)
	engine := gin.New()
	h.Register(engine)
	w := httptest.NewRecorder()
	body := []byte(`{"long_url":"https://example.com/x"}`)
	req, _ := http.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expect 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newFakeRepo()
	ca := &fakeCache{m: map[string]string{}}
	s := service.New(r, ca, 60)
	h := New(s)
	engine := gin.New()
	h.Register(engine)
	w := httptest.NewRecorder()
	body := []byte(`{"long_url":"https://example.com/get"}`)
	req, _ := http.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	var cr createResp
	_ = json.Unmarshal(w.Body.Bytes(), &cr)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/links/"+cr.Code, nil)
	engine.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expect 200, got %d", w.Code)
	}
	type resp struct {
		Code    string `json:"code"`
		LongURL string `json:"long_url"`
	}
	var gr resp
	_ = json.Unmarshal(w.Body.Bytes(), &gr)
	if gr.LongURL != "https://example.com/get" {
		t.Fatalf("expect long_url match, got %s", gr.LongURL)
	}
}

func TestDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newFakeRepo()
	ca := &fakeCache{m: map[string]string{}}
	s := service.New(r, ca, 60)
	h := New(s)
	engine := gin.New()
	h.Register(engine)
	w := httptest.NewRecorder()
	body := []byte(`{"long_url":"https://example.com/del"}`)
	req, _ := http.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	var cr createResp
	_ = json.Unmarshal(w.Body.Bytes(), &cr)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/links/"+cr.Code, nil)
	engine.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Fatalf("expect 204, got %d", w.Code)
	}
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/links/"+cr.Code, nil)
	engine.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Fatalf("expect 404 after delete, got %d", w.Code)
	}
}

func TestBatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newFakeRepo()
	ca := &fakeCache{m: map[string]string{}}
	s := service.New(r, ca, 60)
	h := New(s)
	engine := gin.New()
	h.Register(engine)
	w := httptest.NewRecorder()
	body := []byte(`{"long_urls":["https://a.com","https://b.com","https://a.com"]}`)
	req, _ := http.NewRequest("POST", "/api/links/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expect 200, got %d", w.Code)
	}
	var br batchResp
	_ = json.Unmarshal(w.Body.Bytes(), &br)
	if len(br.Results) != 3 {
		t.Fatalf("expect 3 results, got %d", len(br.Results))
	}
}
