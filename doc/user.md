# 用户模块接口文档

**目录**：

* <a href="#获取用户信息"><b>获取用户信息 -- GET /apidoc/user</b></a>

## <a name="获取用户信息" href="#获取用户信息">获取用户信息</a>

`GET /apidoc/user`

Request header:
- Content-Type: application/json; charset=utf-8

Response header:
- Content-Type: application/json; charset=utf-8

Param - Query

* id (*uint*) 

Return

* code (*int*) 
* msg (*string*) 
* User (*object*) 

Error

* {"code":1,"msg":"认证失败"}
* {"code":2,"msg":"校验失败"}

Example:

<details>
<summary>Param</summary>

```json
id=1
```

</details>

<details>
<summary>Return</summary>

```json
{
    "code": 0,
    "msg": "",
    "User": {
        "id": 1,
        "string": "jd"
    }
}
```

</details>

