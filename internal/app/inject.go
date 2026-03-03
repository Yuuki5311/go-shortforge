package app

import (
	"github.com/ego-component/egorm"
	"github.com/ego-component/eredis"
	"github.com/gotomicro/ego/core/elog"

	"shorturl/internal/cache"
	"shorturl/internal/handler"
	"shorturl/internal/repo"
	"shorturl/internal/service"
)

// 初始化依赖并构建业务服务：
// - 加载 MySQL（Egorm）与 Redis（Eredis）组件
// - 自动迁移表结构
// - 创建缓存与服务实例并返回
func InitDI() (*service.Service, error) {
	gormCmp := egorm.Load("mysql.default").Build()
	redisCmp := eredis.Load("redis.default").Build()
	r := repo.NewGormRepo(gormCmp)
	if err := r.Migrate(); err != nil {
		return nil, err
	}
	c := cache.NewRedisCache(redisCmp)
	s := service.New(r, c, 3600*24)
	return s, nil
}

// 组件初始化入口（由 serve 命令通过 ego.Invoker 调用）：
// - 构建并设置全局业务服务
// - 注册 HTTP 路由到服务器组件
func InitComponents() error {
	s, err := InitDI()
	if err != nil {
		elog.Error("init di failed", elog.FieldErr(err))
		return err
	}
	setService(s)
	h := handler.New(s)
	httpServer.POST("/api/links", h.Create)
	httpServer.GET("/api/links/:code", h.Get)
	httpServer.DELETE("/api/links/:code", h.Delete)
	httpServer.POST("/api/links/batch", h.Batch)
	return nil
}
