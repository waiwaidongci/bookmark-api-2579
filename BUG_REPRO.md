# Bug 复现说明

## Bug 是什么

导出接口按 limit 提前返回部分结果时，没有关闭数据库查询结果集，导致 SQLite 连接被占用；后续请求等待连接，直到超时返回 500。

## 如何触发

1. 创建至少两条书签。
2. 调用 `GET /api/v1/bookmarks/export?limit=1`。
3. 紧接着调用列表接口。

第二次调用会等待约 300ms 后失败，说明连接仍被前一个未关闭的查询占用。

## 错误信息

```text
expected list after export status 200, got 500; body={"code":50000,"data":null,"message":"internal server error"}
```
