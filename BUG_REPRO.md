# Bug 复现说明

## Bug 是什么

新增 URL 已存在的书签时，接口没有返回业务约定中的 409 Conflict，而是返回 500 internal server error。

## 如何触发

1. 启动服务后，向 `POST /api/v1/bookmarks` 发送同一个 URL 的创建请求两次。
2. 第一次返回 201；第二次因 URL 唯一约束触发 SQLite 约束错误。
3. 该错误没有在业务层被转换成重复 URL 错误，最终被处理成 500。

## 错误信息

```text
expected duplicate create status 409, got 500; body={"code":50000,"data":null,"message":"internal server error"}
```
