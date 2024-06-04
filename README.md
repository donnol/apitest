# apitest

API test for Go.

在测试的同时生成接口文档。

确保代码和文档的一致。

省心又省力。

## 如何使用

1. 添加接口

1.1. 定义接口、参数和结果

`POST /book`

```go
type BookParam struct {
    Name string `json:"name"` // 名称
    Author string `json:"author"` // 作者
}

type BookResult struct {
    Id uint64 `json:"id"` // 新记录id
}

func CreateBook(c *gin.Context) {
    // use BookParam receive param
    var p BookParam

    // return BookResult as API's result
    var r BookResult
    c.JSON(http.StatusOK, r)
}
```

1.2. 注册路由，启动服务：

```go
func main() {
	buf := new(bytes.Buffer)
	gin.DefaultWriter = io.MultiWriter(buf, os.Stdout)
	engine := gin.Default()
    engine.POST("/book", CreateBook)

	if err := engine.Run(":8888"); err != nil {
        panic(err)
    }
}
```

2. 添加测试

```go
func TestAPI(t *testing.T) {
    w := new(bytes.Buffer)

    var r BookResult

    // 调用请求的同时生成文档
    if err := NewAT("/book", http.MethodPost, "新建图书信息", nil, nil). // 新建测试，指定接口路径和方法
        SetParam(&BookParam{ // 设置参数
            Name: "test",
            Author: "jd",
        }).
        FakeRun().  // 虚假执行；若要真正执行，请使用`Run()`
        Result(&r). // 获取结果
        WriteFile(w). // 输出文档到文件
        Err(); err != nil {
        t.Fatal(err)
    }
}
```

3. 执行测试并生成文档

`go test .`

将生成文档：

```md
## 新建图书信息

`POST /book`

Param - Query

* name (*string*) 名称
* author (*string*) 作者

Return

* id (*uint64*) 新记录id
```

其中，参数和返回的字段信息来源于结构体的定义。

如此，在接口和测试完成的同时，文档也随之完成。
