# Bug 复现说明

## Bug 是什么

列表查询没有把调用方传入的 context 取消状态传递给数据库操作，导致已经取消的请求仍继续查询并返回成功。

## 如何触发

1. 创建一条书签。
2. 取消用于列表查询的 context。
3. 调用列表查询。

预期应返回 `context.Canceled`，实际返回成功列表，且没有取消传播。

## 错误信息

```text
expected context.Canceled, got <nil>
```
