# Bug 复现说明

## Bug 是什么

标签统计接口在书签包含标签时，向 nil map 写入导致 panic，最终返回 500；没有标签时返回的 tags 也可能是 null 而不是空数组。

## 如何触发

1. 创建一个带 `go,docs` 标签的书签。
2. 调用 `GET /api/v1/bookmarks/stats/tags`。
3. 服务触发 panic，恢复中间件将其转换为 500。

## 错误信息

```text
panic_recovered error="assignment to entry in nil map"
expected tag stats status 200, got 500; body={"code":50000,"data":null,"message":"internal server error"}
```
