package apitest

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/donnol/do"
	"github.com/gin-gonic/gin"
	"github.com/jaswdr/faker"
	"github.com/samber/lo"
)

const (
	authHeaderKey         = "Authorization"
	authHeaderValuePrefix = "Bearer "
	authHeaderSign        = "Signature"
)

type Collector struct {
	testAPIs    map[string]*TestAPI
	testAPIKeys []string
}

func NewCollector(
	obj interface {
		RegisterAPI(apiGroup *gin.RouterGroup) []*Route
	},
	routem map[string]lo.Tuple2[reflect.Value, int],
) *Collector {
	collector := &Collector{}

	// key is apiKey2(method, fullpath)
	//
	// use `findTestAPIsByPrefix` to get api batch by prefix
	// test engine
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()
	apiGroup := engine.Group("/api")

	// test api
	if routem == nil {
		routem = make(map[string]lo.Tuple2[reflect.Value, int])
	}
	routes := obj.RegisterAPI(apiGroup)
	// fmt.Printf("routem: %+v\n", routem)

	// generate apitest object
	pathKeys := make([]string, 0, 64)
	m := make(map[string]*TestAPI)
	pathm := make(map[string]int)
	setAPI := func(basePath string, route *Route) {
		pathkey := apiKey2(route.Method, route.Path)
		v, ok := pathm[pathkey]
		if ok {
			pathm[pathkey] = v + 1
		} else {
			pathm[pathkey] = 1
		}

		var param, result reflect.Type
		routeKey := apiKey2(route.Method, route.Path, v)
		if tval, ok := routem[routeKey]; ok {
			// fmt.Printf("route key with param: %s\n", routeKey)
			val, count := tval.Unpack()
			_ = count
			ptyp := val.Type().In(1)
			rtyp := val.Type().Out(0)

			param = ptyp
			result = rtyp
		}
		// fullPath肯定是唯一的，但route.Path则不是：
		//  route.Path是/page的情况下，fullPath可能是/user/page，也可能是/book/page
		fullPath := basePath + route.Path
		key := apiKey2(route.Method, fullPath)
		pathKeys = append(pathKeys, key)
		at := NewAT(fullPath, route.Method, route.Comment, nil, nil)
		if route.Opt.NeedLogin {
			at.MarkAuthHeader(authHeaderKey, "Bearer [TOKEN]")
		}
		if route.Opt.ParamFormat == "xml" {
			at.UseXMLParamFormat()
		}
		if route.Opt.ResultFormat == "xml" {
			at.UseXMLResultFormat()
		}
		m[key] = &TestAPI{
			AT:     at,
			key:    key,
			param:  param,
			result: result,
		}
	}

	HandleRoutes(apiGroup, routes, func(group *gin.RouterGroup, route *Route) {
		setAPI(group.BasePath(), route)
	})

	collector.testAPIs = m
	collector.testAPIKeys = pathKeys

	return collector
}

type Route = do.Route[*gin.Context]

func HandleRoutes(group *gin.RouterGroup, routes []*Route, h func(group *gin.RouterGroup, route *Route)) {
	for _, route := range routes {
		route := route
		if route == nil {
			continue
		}

		if route.Handler != nil {
			h(group, route)
		}
		if len(route.Childs) > 0 {
			childGroup := group.Group(route.Path)
			HandleRoutes(childGroup, route.Childs, h)
		}
	}
}

type TestAPI struct {
	*AT

	key           string
	param, result reflect.Type
}

// getParamResult 获取key所对应的参数和结果
func (t *TestAPI) GetParamResult(gen ...func(key string, param, result reflect.Type) (p, r any)) (p, r any) {
	if len(gen) > 0 && gen[0] != nil {
		return gen[0](t.key, t.param, t.result)
	}
	return GenParamResult(t.key, t.param, t.result)
}

func GenParamResult(key string, param, result reflect.Type) (p, r any) {
	// 默认使用随机值
	p = fillStruct(key, param)
	rv := fillStruct(key, result)
	r = Result[any]{Data: rv}

	return
}

func (c *Collector) TestAPIs() map[string]*TestAPI {
	return c.testAPIs
}

func (c *Collector) TestAPIKeys() []string {
	return c.testAPIKeys
}

func (c *Collector) FindTestAPIsByPrefix(prefix string) (r []*TestAPI) {
	for _, key := range c.testAPIKeys {
		item := c.testAPIs[key]

		key := apiKey2(item.Method(), item.Path())
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		r = append(r, item)
	}

	return
}

func fillStruct(key string, v reflect.Type) any {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("fill %q data failed: %+v, v'is type: %v\n", key, r, v)
		}
	}()

	if v != nil {
		s := reflect.New(v).Interface()
		faker.New().Struct().Fill(s)
		return s
	}

	return nil
}

type Result[T any] struct {
	Code      int    `json:"code"`                // 业务码，0为正常，非0表示出错
	Msg       string `json:"msg"`                 // 出错信息
	Timestamp int64  `json:"timestamp,omitempty"` // 时间戳
	TraceId   string `json:"traceId"`             // 追踪id

	Data T `json:"data"` // 业务数据
}

func apiKey2(method, path string, i ...int) string {
	if len(i) == 0 || i[0] == 0 {
		return path + " " + method
	}
	return path + " " + method + " " + strconv.Itoa(i[0])
}