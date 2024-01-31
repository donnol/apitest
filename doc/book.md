# 图书模块接口文档

**目录**：

* <a href="#获取图书信息"><b>获取图书信息 -- GET /apidoc/book</b></a>

## <a name="获取图书信息" href="#获取图书信息">获取图书信息</a>

`GET /apidoc/book`

Request header:
- Content-Type: application/json; charset=utf-8
- : 

Response header:
- Content-Type: application/json; charset=utf-8

Param - Query

* id (*uint*) 

Return

* code (*int*) 
* msg (*string*) 
* Book (*object*) 

Error

* {"code":1,"msg":"认证失败"}
* {"code":2,"msg":"校验失败"}


<details>
<summary>Try to run</summary>
<div>
<div>
<label for="Params(参照下面的示例)"><a href="">Params(参照下面的示例)</a></label>
<p></p>
<textarea rows="4" cols="50" name="Params(参照下面的示例)" id="param/apidoc/book GET" placeholder='id=1'>id=1</textarea>
</div>
<div>
<button onclick="sendRequest('get', '/apidoc/book', 'token/apidoc/book GET', 'param/apidoc/book GET', 'result/apidoc/book GET')">Try to run</button>
<pre id="result/apidoc/book GET" style="font-size: large"></pre>
</div>
</div>
</details>


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
    "Book": {
        "id": 1,
        "string": "jd",
        "page": 100
    }
}
```

</details>

