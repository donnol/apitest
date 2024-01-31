# 用户接口文档

**目录**：

* <a href="#获取用户信息"><b>获取用户信息 -- GET /api/user</b></a>

* <a href="#添加用户信息"><b>添加用户信息 -- POST /api/user</b></a>

* <a href="#导入用户信息(以csv文件格式)"><b>导入用户信息(以csv文件格式) -- POST /api/user/import</b></a>

## <a name="获取用户信息" href="#获取用户信息">获取用户信息</a>

`GET /api/user`

Request header:
- Content-Type: application/json; charset=utf-8
- : 

Response header:
- Content-Type: application/json; charset=utf-8

Param - Query

* id (*string*) id
* name (*string*) 名字
* age (*int*) 年龄
* addr (*object*) 地址
    * city (*string*) 城市
    * home (*string*) 家
* phone (*string*) 手机

Return

* id (*string*) id
* name (*string*) 名字
* age (*int*) 年龄
* addr (*object*) 地址
    * city (*string*) 城市
    * home (*string*) 家
* phone (*string*) 手机


<details>
<summary>Try to run</summary>
<div>
<div>
<label for="Params(参照下面的示例)"><a href="">Params(参照下面的示例)</a></label>
<p></p>
<textarea rows="4" cols="50" name="Params(参照下面的示例)" id="param/api/user GET" placeholder='addr=%7B+%7D&age=0&id=1&name=jd&phone='>addr=%7B+%7D&age=0&id=1&name=jd&phone=</textarea>
</div>
<div>
<button onclick="sendRequest('get', '/api/user', 'token/api/user GET', 'param/api/user GET', 'result/api/user GET')">Try to run</button>
<pre id="result/api/user GET" style="font-size: large"></pre>
</div>
</div>
</details>


Example:

<details>
<summary>Param</summary>

```json
addr=%7B+%7D&age=0&id=1&name=jd&phone=
```

</details>

<details>
<summary>Return</summary>

```json
{
    "id": "1", // id
    "name": "jd", // 名字
    "age": 0, // 年龄
    "addr": { // 地址
        "city": "", // 城市
        "home": "" // 家
    },
    "phone": "" // 手机
}
```

</details>

## <a name="添加用户信息" href="#添加用户信息">添加用户信息</a>

`POST /api/user`

Request header:
- Content-Type: application/json; charset=utf-8
- : 

Response header:
- Content-Type: application/json; charset=utf-8

Param - Body

* id (*string*) id
* name (*string*) 名字
* age (*int*) 年龄
* addr (*object*) 地址
    * city (*string*) 城市
    * home (*string*) 家
* phone (*string*) 手机

Return

* id (*string*) id
* name (*string*) 名字
* age (*int*) 年龄
* addr (*object*) 地址
    * city (*string*) 城市
    * home (*string*) 家
* phone (*string*) 手机


<details>
<summary>Try to run</summary>
<div>
<div>
<label for="Params(参照下面的示例)"><a href="">Params(参照下面的示例)</a></label>
<p></p>
<textarea rows="4" cols="50" name="Params(参照下面的示例)" id="param/api/user POST" placeholder='{"id":"1","name":"jd","age":0,"addr":{"city":"","home":""},"phone":""}'>{"id":"1","name":"jd","age":0,"addr":{"city":"","home":""},"phone":""}</textarea>
</div>
<div>
<button onclick="sendRequest('post', '/api/user', 'token/api/user POST', 'param/api/user POST', 'result/api/user POST')">Try to run</button>
<pre id="result/api/user POST" style="font-size: large"></pre>
</div>
</div>
</details>


Example:

<details>
<summary>Param</summary>

```json
{
    "id": "1", // id
    "name": "jd", // 名字
    "age": 0, // 年龄
    "addr": { // 地址
        "city": "", // 城市
        "home": "" // 家
    },
    "phone": "" // 手机
}
```

</details>

<details>
<summary>Return</summary>

```json
{
    "id": "1", // id
    "name": "jd", // 名字
    "age": 0, // 年龄
    "addr": { // 地址
        "city": "", // 城市
        "home": "" // 家
    },
    "phone": "" // 手机
}
```

</details>

## <a name="导入用户信息(以csv文件格式)" href="#导入用户信息(以csv文件格式)">导入用户信息(以csv文件格式)</a>

`POST /api/user/import`

Request header:
- Content-Type: text/csv; charset=utf-8
- : 

Response header:
- Content-Type: application/json; charset=utf-8

Param - Body

* id (*string*) id
* name (*string*) 名字
* age (*int*) 年龄
* addr (*object*) 地址
    * city (*string*) 城市
    * home (*string*) 家
* phone (*string*) 手机

Return

* id (*string*) id
* name (*string*) 名字
* age (*int*) 年龄
* addr (*object*) 地址
    * city (*string*) 城市
    * home (*string*) 家
* phone (*string*) 手机


<details>
<summary>Try to run</summary>
<div>
<div>
<label for="Params(参照下面的示例)"><a href="">Params(参照下面的示例)</a></label>
<p></p>
<textarea rows="4" cols="50" name="Params(参照下面的示例)" id="param/api/user/import POST" placeholder='{"id":"1","name":"jd","age":0,"addr":{"city":"","home":""},"phone":""}'>{"id":"1","name":"jd","age":0,"addr":{"city":"","home":""},"phone":""}</textarea>
</div>
<div>
<button onclick="sendRequest('post', '/api/user/import', 'token/api/user/import POST', 'param/api/user/import POST', 'result/api/user/import POST')">Try to run</button>
<pre id="result/api/user/import POST" style="font-size: large"></pre>
</div>
</div>
</details>


Example:

<details>
<summary>Param</summary>

```json
{
    "id": "1", // id
    "name": "jd", // 名字
    "age": 0, // 年龄
    "addr": { // 地址
        "city": "", // 城市
        "home": "" // 家
    },
    "phone": "" // 手机
}
```

</details>

<details>
<summary>Return</summary>

```json
{
    "id": "1", // id
    "name": "jd", // 名字
    "age": 0, // 年龄
    "addr": { // 地址
        "city": "", // 城市
        "home": "" // 家
    },
    "phone": "" // 手机
}
```

</details>

