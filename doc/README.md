# 接口文档说明

## 认证

先调用`login`接口，获得`token`。

在后续接口调用时在`header`上添加：`Authorization`: `Bearer [token]`.

如：
```sh
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDM5OTQwODksImlkIjoiNDkxNTgyMDczOTM3NjI2Nzc1Iiwib2lkIjoiMSIsInRhZyI6InVuaWZpZWRfbG9naW4iLCJ0aWQiOiIxMDAxIn0.8oMA2WJ3bQmWOVcGofaGBIg3vwWOyQtpJc6Dh2he3ao
```
