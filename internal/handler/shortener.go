package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"shorturl/internal/service"
)

// 短链处理器：负责 HTTP 层的参数校验、状态码与响应构造
type ShortenerHandler struct {
	svc *service.Service
}

// 创建处理器实例（持有业务服务）
func New(svc *service.Service) *ShortenerHandler {
	return &ShortenerHandler{svc: svc}
}

// 创建短链请求体
type createReq struct {
	LongURL string `json:"long_url" binding:"required,url"`
}
// 创建短链响应体
type createResp struct {
	Code     string `json:"code"`
	LongURL  string `json:"long_url"`
	ShortURL string `json:"short_url"`
}

// 批量创建请求体
type batchReq struct {
	LongURLs []string `json:"long_urls" binding:"required"`
}
// 批量创建响应体（按输入顺序返回结果）
type batchResp struct {
	Results []createResp `json:"results"`
}

// 注册 /api/links 路由分组
func (h *ShortenerHandler) Register(r *gin.Engine) {
	g := r.Group("/api/links")
	g.POST("", h.Create)
	g.GET(":code", h.Get)
	g.DELETE(":code", h.Delete)
	g.POST("batch", h.Batch)
}

// 创建短链：
// - 校验请求体
// - 调用业务服务生成短码（幂等）
// - 返回包含短链的完整 URL
func (h *ShortenerHandler) Create(c *gin.Context) {
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	code, err := h.svc.Shorten(c.Request.Context(), req.LongURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, createResp{Code: code, LongURL: req.LongURL, ShortURL: shortURL(c, code)})
}

// 查询短链：
// - 通过短码解析长链
// - 未命中返回 404
func (h *ShortenerHandler) Get(c *gin.Context) {
	code := c.Param("code")
	long, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if long == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"long_url": long, "code": code})
}

// 删除短链：
// - 删库并删除缓存键
// - 返回 204 无内容
func (h *ShortenerHandler) Delete(c *gin.Context) {
	code := c.Param("code")
	if err := h.svc.Delete(c.Request.Context(), code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// 批量创建短链：
// - 输入去重
// - 已存在复用短码，未存在批量生成并入库
func (h *ShortenerHandler) Batch(c *gin.Context) {
	var req batchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mp, err := h.svc.BatchShorten(c.Request.Context(), req.LongURLs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := batchResp{Results: make([]createResp, 0, len(mp))}
	for _, u := range req.LongURLs {
		if code, ok := mp[u]; ok {
			resp.Results = append(resp.Results, createResp{Code: code, LongURL: u, ShortURL: shortURL(c, code)})
		}
	}
	c.JSON(http.StatusOK, resp)
}

// 组装短链完整 URL（根据请求的协议与主机）
func shortURL(c *gin.Context, code string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + "/u/" + code
}
