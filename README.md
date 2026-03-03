短链接 Demo（Ego + Egorm + Eredis + Cobra）

概览
- 技术栈：Ego + Egin（HTTP）/ Egorm（MySQL）/ Eredis（Redis）/ Cobra（启动命令）
- 功能：长链生成短链、查询、删除、批量生成；提供跳转端点 /u/:code
- 特性：幂等短链生成、Redis 缓存加速、自动迁移、完善测试、ab 压测脚本

目录结构（关键文件）
- 入口与命令：main.go、cmd/root.go、cmd/serve.go
- 装配与路由：internal/app/server.go、internal/app/inject.go
- 业务与存储：internal/service/shortener.go、internal/repo/repo.go、internal/model/model.go
- 缓存：internal/cache/{cache.go,redis.go}
- HTTP：internal/handler/shortener.go
- 配置与脚本：config.toml、docker-compose.yml、docs/ab.md、scripts/payloads/*

快速开始
1) 准备环境
- 安装 Go（1.20+）
- 可选：安装 Docker Desktop（推荐用于一键启动 MySQL/Redis）

2) 获取依赖并构建
- go mod tidy

3) 启动依赖
- 使用 Docker（推荐）：
  - 在项目根目录执行：docker compose up -d
  - 默认创建（若本机占用 3306，已映射为 3307）：
    - MySQL：127.0.0.1:3307，root/123456789Lin，数据库 ego_short
    - Redis：127.0.0.1:6379（无密码）
- 或者使用自有 MySQL/Redis，修改 config.toml 中的 DSN/地址

4) 启动服务
- go run . --config=config.toml serve
- 环境变量可选：EGO_DEBUG=true（输出组件请求详情）

API 文档
- POST /api/links 创建单个短链
  - 请求：
    - Content-Type: application/json
    - {"long_url":"https://example.com/path?a=1"}
  - 响应：
    - 200: {"code":"abc1234","long_url":"...","short_url":"http://127.0.0.1:8080/u/abc1234"}
- GET /api/links/:code 查询长链
  - 响应：
    - 200: {"code":"abc1234","long_url":"..."}
    - 404: {"error":"not found"}
- DELETE /api/links/:code 删除短链
  - 响应：
    - 204: 空
- POST /api/links/batch 批量创建
  - 请求：
    - {"long_urls":["https://a.com","https://b.com/1","https://a.com?q=1"]}
  - 响应：
    - 200: {"results":[{"code":"...","long_url":"...","short_url":"..."}]}
- GET /u/:code 跳转重定向（辅助）
  - 302 重定向到长链

配置说明（config.toml）
- server.http：服务端口与模式（release）
- mysql.default：Egorm DSN（默认 root:123456789Lin@tcp(127.0.0.1:3307)/ego_short）
- redis.default：Eredis 连接参数（addr/db/超时）
- 如端口/密码变更，请同步修改 docker-compose.yml 与 config.toml

实现要点（代码位置）
- 短链生成（幂等）：internal/service/shortener.go
  - 先按 long_url 查复用短码 → 否则生成 7 位 base62 短码（最多 5 次避免冲突）→ 入库 → 写缓存
- 查询与缓存：internal/service/shortener.go
  - 优先 Redis → 未命中读库并回填 → 返回
- 删除：internal/service/shortener.go
  - 删库 → 删缓存键
- 批量生成：internal/service/shortener.go
  - 输入去重 → 已存在复用 → 未存在批量生成/批量入库/批量回填
- 存储（Egorm）：internal/repo/repo.go、internal/model/model.go
  - AutoMigrate 表 short_urls，code 与 long_url 唯一索引
- 缓存（Eredis）：internal/cache/redis.go
  - key=shorturl:code:{code}，TTL 默认 24h
- 路由（Egin/Gin）：internal/app/inject.go、internal/handler/shortener.go
  - /api/links、/api/links/:code、/api/links/batch、/u/:code
- 启动（Cobra + Ego）：cmd/serve.go、internal/app/server.go

测试
- 单元测试：go test ./...
  - 服务层：internal/service/shortener_test.go（覆盖生成/查询/删除/批量）
  - HTTP 层：internal/handler/shortener_test.go（验证 POST /api/links）
- 提示：测试使用内存替身，不依赖外部 DB/Redis，执行快速稳定

压测（ab）
- 文档与命令：docs/ab.md
- 示例（需已启动服务）：
  - 单个生成：ab -n 1000 -c 50 -p scripts\payloads\single.json -T application/json http://127.0.0.1:8080/api/links
  - 查询：ab -n 2000 -c 100 http://127.0.0.1:8080/api/links/<CODE>
  - 删除：ab -n 500 -c 50 -m DELETE http://127.0.0.1:8080/api/links/<CODE>
  - 批量：ab -n 200 -c 20 -p scripts\payloads\batch.json -T application/json http://127.0.0.1:8080/api/links/batch
- Windows 可通过 WSL 或安装 httpd 工具包使用 ab
