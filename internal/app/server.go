package app

import (
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/server/egin"

	"shorturl/internal/service"
)

var (
	httpServer *egin.Component
	svc        *service.Service
)

// 构建 HTTP 服务器组件：
// - 创建 Gin 引擎并挂载基础中间件（恢复、访问日志）
// - 载入 Egin 配置 server.http 并生成组件
// - 预注册跳转端点 /u/:code
func NewHTTPServer() *egin.Component {
	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger())
	httpServer = egin.Load("server.http").Build()
	httpServer.GET("/u/:code", redirect)
	return httpServer
}

// 设置业务服务实例（供路由处理与跳转端点使用）
func setService(s *service.Service) { svc = s }

// 跳转端点：根据短码解析长链并重定向
// - 服务未就绪返回 500
// - 未命中返回 404
// - 命中则返回 302 并跳转到长链
func redirect(c *gin.Context) {
	if svc == nil {
		c.JSON(500, gin.H{"error": "service not ready"})
		return
	}
	code := c.Param("code")
	long, err := svc.Resolve(c.Request.Context(), code)
	if err != nil || long == "" {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.Redirect(302, long)
}
