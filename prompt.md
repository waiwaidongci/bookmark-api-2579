# 项目：收藏链接API

从0做一个Go收藏链接API，用Gin开发，数据存SQLite。链接记录包含标题、URL、标签、备注、点击次数和创建时间，支持新增修改删除、按标签筛选、关键词搜索、记录点击次数、检查重复URL。代码遵循Go企业分层规范：cmd/server/main.go、internal/config、internal/model、internal/repository、internal/service、internal/handler、internal/router、internal/middleware、migrations。数据库写入统一通过repository层，URL校验和重复检查放在service层，列表接口支持分页。
