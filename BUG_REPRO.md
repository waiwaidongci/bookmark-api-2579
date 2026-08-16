# Bug 复现说明

## Bug 是什么

批量删除接口在 ID 列表包含不存在记录时，仍返回 200，并执行部分删除；正确行为应返回 404 且整批回滚。

## 如何触发

1. 创建两条书签。
2. 调用 `POST /api/v1/bookmarks/batch-delete`，请求体包含 `{"ids":[1,999]}`。
3. 接口返回 200，并删除了 ID 为 1 的记录；ID 999 被静默忽略。

## 错误信息

```text
expected missing-id batch delete status 404, got 200; body={"code":0,"data":{"deleted":1},"message":"success"}
```
