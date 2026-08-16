# Bookmark API

一个使用 Gin 和 SQLite 实现的收藏链接 API，遵循 Go 企业分层结构。

## 目录结构

```text
cmd/server/         服务入口
internal/config     配置加载
internal/model      数据模型
internal/repository SQLite 数据访问
internal/service    业务规则：URL 校验、重复检查、分页归一化
internal/handler    HTTP 处理器
internal/router     路由注册
internal/middleware 日志与 panic 恢复
migrations          SQLite 迁移脚本
```

## 运行

```bash
go mod tidy
go run ./cmd/server
```

默认监听 `18007` 端口，数据文件写入 `./data/bookmarks.db`。

可用环境变量：

- `PORT`：HTTP 端口，默认 `18007`
- `DB_PATH`：SQLite 文件路径，默认 `./data/bookmarks.db`
- `GIN_MODE`：Gin 运行模式，默认 `release`

## API

统一响应结构：

```json
{"code":0,"message":"success","data":{}}
```

### 新增

```http
POST /api/v1/bookmarks
Content-Type: application/json

{
  "title": "Go documentation",
  "url": "https://go.dev/doc/",
  "tags": "go, docs",
  "note": "official docs"
}
```

### 列表

```http
GET /api/v1/bookmarks?page=1&page_size=20&tag=go&keyword=sqlite
```

`page_size` 最大为 `100`。`tag` 按标签精确匹配，`keyword` 搜索标题、URL、备注和标签。

### 详情

```http
GET /api/v1/bookmarks/:id
```

### 修改

```http
PUT /api/v1/bookmarks/:id
Content-Type: application/json

{
  "title": "Go docs",
  "note": "updated"
}
```

字段均可选，至少提供一个字段。

### 记录点击

```http
POST /api/v1/bookmarks/:id/click
```

### 删除

```http
DELETE /api/v1/bookmarks/:id
```

### 健康检查

```http
GET /healthz
```

## 测试

```bash
go test ./...
```
