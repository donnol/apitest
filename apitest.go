// Package apitest Usage:
//
//	NewAT(xxx).
//		SetParam(xxx).
//		Debug().
//		Run().
//		EqualCode(xxx).
//		Result(xxx).
//		Equal(...).
//		WriteFile(xxx).
//		Err()
package apitest

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/donnol/do"
)

// Predefined error
var (
	// ErrNilParam 参数为nil
	ErrNilParam = errors.New("please input param, param is nil now")
)

// AT api test
type AT struct {
	// 服务器配置
	scheme             string
	host               string
	port               string
	url                url.URL
	caCertPath         string
	certFile           string
	keyFile            string
	insecureSkipVerify bool

	// 请求相关
	authHeaderKey   string
	authHeaderValue string
	path            string
	method          string
	comment         string
	clientTimeout   time.Duration
	header          http.Header
	cookies         []*http.Cookie
	param           any
	paramFormat     string // 结果格式，默认为`json`
	file            string // 文件
	resultWrapper   ResultWrapper
	result          any
	resultFormat    string // 结果格式，默认为`json`
	ates            []any
	handlerMap      map[string]any // 如："gin.HandlerFunc", gin.HandlerFunc(nil),

	// 请求和响应
	req     *http.Request
	reqBody []byte
	resp    *http.Response

	// 接口实现情况
	status Status

	// 文档
	doc string

	// 调试
	debug bool

	// 慢请求数量
	slowNum int

	// 是否批量压力测试中
	isPressureBatch bool

	err error
}

type ResultWrapper interface {
	WithData(data any)
}

// NewAT 新建
func NewAT(
	path,
	method,
	comment string,
	h http.Header,
	cookies []*http.Cookie,
) *AT {
	at := &AT{
		path:    path,
		method:  method,
		comment: comment,
		header:  h,
		cookies: cookies,
	}

	_, err := url.Parse(path)
	if err != nil {
		at.setErr(err)
	}

	return at
}

// New 克隆一个新的AT
func (at *AT) New() *AT {
	return at.clone()
}

// SetScheme 设置scheme
func (at *AT) SetScheme(scheme string) *AT {
	at.scheme = scheme
	return at
}

// SetClientTimeout set client timeout like: 10*time.Second
func (at *AT) SetClientTimeout(timeout time.Duration) *AT {
	at.clientTimeout = timeout
	return at
}

func (at *AT) SetCert(caCertPath, certFile, keyFile string) *AT {
	at.caCertPath = caCertPath
	at.certFile = certFile
	at.keyFile = keyFile
	return at
}

func (at *AT) InsecureSkipVerify() *AT {
	at.insecureSkipVerify = true
	return at
}

// SetHost 设置host
func (at *AT) SetHost(host string) *AT {
	at.host = host
	return at
}

// SetPort 设置端口，如":8080"
func (at *AT) SetPort(port string) *AT {
	at.port = port
	return at
}

// SetHeader 设置header
func (at *AT) SetHeader(header http.Header) *AT {
	at.header = header
	return at
}

func (at *AT) MarkAuthHeader(authHeaderKey, authHeaderValue string) *AT {
	at.authHeaderKey = authHeaderKey
	at.authHeaderValue = authHeaderValue
	return at
}

// SetCookies 设置cookies
func (at *AT) SetCookies(cookies []*http.Cookie) *AT {
	at.cookies = cookies
	return at
}

// SetParam 设置参数
func (at *AT) SetParam(param any) *AT {
	if param == nil {
		at.setErr(fmt.Errorf("nil param"))
		return at
	}

	at.param = param
	return at
}

// SetFile 设置文件
func (at *AT) SetFile(file string) *AT {
	if file == "" {
		at.setErr(fmt.Errorf("empty file"))
		return at
	}

	at.file = file
	return at
}

type (
	Status int
)

const (
	StatusNone           Status = 0
	StatusInDesign       Status = 1 // 设计中
	StatusNotImplemented Status = 2 // 未实现
	StatusImplementation Status = 3 // 实现中
	StatusImplemented    Status = 4 // 已实现
)

func (s Status) String() string {
	r := ""
	switch s {
	case StatusInDesign:
		r = "设计中"
	case StatusNotImplemented:
		r = "未实现"
	case StatusImplementation:
		r = "实现中"
	case StatusImplemented:
		r = "已实现"
	}
	return r
}

func (at *AT) SetStatus(status Status) *AT {
	at.status = status
	return at
}

// UseXMLFormat 设置参数和结果格式为XML
func (at *AT) UseXMLFormat() *AT {
	at.paramFormat = "xml"
	at.resultFormat = "xml"
	return at
}

// UseXMLParamFormat 设置参数格式为XML
func (at *AT) UseXMLParamFormat() *AT {
	at.resultFormat = "xml"
	return at
}

// UseXMLResultFormat 设置结果格式为XML
func (at *AT) UseXMLResultFormat() *AT {
	at.resultFormat = "xml"
	return at
}

// Run 运行
func (at *AT) Run() *AT {
	return at.run(true)
}

// Run 运行
func (at *AT) FakeRun() *AT {
	return at.run(false)
}

// MonkeyRun 猴子运行
func (at *AT) MonkeyRun() *AT {
	if at.param == nil {
		at.setErr(ErrNilParam)
		return at
	}

	// 根据参数结构体随机生成测试值
	if err := gofakeit.Struct(at.param); err != nil {
		at.setErr(fmt.Errorf("generate param value failed: %w", err))
		return at
	}
	at.jsonIndent(os.Stdout, at.param)

	return at.run(true)
}

// PressureRun 压力运行，n: 运行次数，c: 并发数
func (at *AT) PressureRun(n, c int) *AT {
	w := do.NewWorker(c)
	w.Start()

	// 记录开始时间
	before := time.Now()

	var total int64
	for i := 0; i < n; i++ {
		if err := w.Push(*do.NewJob(func(ctx context.Context) error {
			// 运行
			at.run(true)

			// 统计数量
			atomic.AddInt64(&total, 1)

			return nil
		}, 0, nil)); err != nil {
			at.setErr(err)
		}
	}

	w.Stop()

	// 记录结束时间，并计算耗时
	used := time.Since(before)
	avg := float64(total) / float64(used.Milliseconds()) * 1000
	fmt.Printf("\n=== Pressure Report ===\nNumber: %d\nConcurrency: %d\nCompleted: %d\nUsed time: %vs\nRPS: %v\n=== END ===\n\n", n, c, total, do.Round(used.Seconds(), 2), do.Round(avg, 2))

	return at
}

// PressureParam 压力测试参数
type PressureParam struct {
	N int // 运行次数
	C int // 并发数
}

// PressureRunBatch 批量压力运行
func (at *AT) PressureRunBatch(param []PressureParam) *AT {
	at.isPressureBatch = true

	for _, single := range param {
		at = at.PressureRun(single.N, single.C)
	}

	fmt.Printf("slowNum is %d\n", at.slowNum)
	at.isPressureBatch = false

	return at
}

// Debug 开启调试模式
func (at *AT) Debug() *AT {
	at.debug = true
	return at
}

// EqualCode 比较响应码
func (at *AT) EqualCode(wantCode int) *AT {
	// 复制resp.Body数据
	data, _, err := copyResponseBody(at.resp)
	if err != nil {
		at.setErr(err)
		return at
	}

	// 校验响应码
	if at.resp.StatusCode == wantCode {
		return at
	}

	at.setErr(fmt.Errorf("bad status code, got %+v\ndata is %s", at.resp, data))
	return at
}

var (
	resultExtractor = make(map[string]ResultExtractor)
)

type (
	ResultExtractor func(data []byte, r any) error
)

func (at *AT) RegisterResultExtractor(format string, re ResultExtractor) *AT {
	resultExtractor[format] = re
	return at
}

func (at *AT) GetResultExtractor(format string) (re ResultExtractor, ok bool) {
	re, ok = resultExtractor[format]
	return re, ok
}

func extract(format string, data []byte, r any) error {
	re, ok := resultExtractor[format]
	if ok {
		return re(data, r)
	}

	switch format {
	case "xml":
		if err := xml.Unmarshal(data, r); err != nil {
			return fmt.Errorf("xml decode failed: %+v, data: %s", err, data)
		}
	default:
		if err := json.Unmarshal(data, r); err != nil {
			return fmt.Errorf("json decode failed: %+v, data: %s", err, data)
		}
	}
	return nil
}

// ResultWrapper 指定结果包装结构
func (at *AT) ResultWrapper(rw ResultWrapper) *AT {
	if rw == nil {
		at.setErr(fmt.Errorf("result wrapper r can't be nil"))
		return at
	}

	at.resultWrapper = rw

	return at
}

// Result 获取结果
func (at *AT) Result(r any) *AT {
	if r == nil {
		at.setErr(fmt.Errorf("result r can't be nil"))
		return at
	}

	// 复制resp.Body
	if at.resp != nil {
		data, _, err := copyResponseBody(at.resp)
		if err != nil {
			at.setErr(fmt.Errorf("copy response body failed: %+v, resp: %+v", err, at.resp))
			return at
		}

		// 解析data到r
		if err := extract(at.resultFormat, data, r); err != nil {
			at.setErr(err)
			return at
		}
	}

	// 当有resultWrapper时，使用它来包装结果
	if at.resultWrapper != nil {
		at.resultWrapper.WithData(r)
		at.result = at.resultWrapper
	} else {
		at.result = r
	}

	at.jsonIndent(os.Stdout, r)

	return at
}

// Errors 获取错误
func (at *AT) Errors(errs ...any) *AT {
	at.ates = errs
	return at
}

// Equal 校验
func (at *AT) Equal(args ...any) *AT {
	l := len(args)
	d := l % 2
	if d != 0 {
		at.setErr(fmt.Errorf("please Input Double Args: %v", args))
		return at
	}
	for i := 0; i < l; i += 2 {
		if !reflect.DeepEqual(args[i], args[i+1]) {
			at.setErr(fmt.Errorf("no.%d Not Equal, Have %v, Want %v", i/2+1, args[i], args[i+1]))
			return at
		}
	}

	return at
}

// EqualThen 相等之后
func (at *AT) EqualThen(f func(*AT) error, args ...any) *AT {
	// 先比较args
	at = at.Equal(args...)
	if at.err != nil {
		return at
	}

	// 成功之后才继续运行f
	if err := f(at); err != nil {
		at.setErr(err)
		return at
	}

	return at
}

// WriteFile 写入markdown文件
func (at *AT) WriteFile(w io.Writer) *AT {
	if w == nil {
		at.setErr(fmt.Errorf("nil writer"))
		return at
	}

	if at.doc == "" {
		at.makeDoc() // 尝试一次生成文档
	}

	if at.doc == "" {
		at.setErr(fmt.Errorf("empty doc"))
		return at
	}

	if _, err := w.Write([]byte(at.doc)); err != nil {
		at.setErr(err)
		return at
	}
	return at
}

// === Get api info ===

func (at *AT) Title() string {
	return at.comment
}

func (at *AT) Method() string {
	return at.method
}

func (at *AT) Path() string {
	return at.path
}

func (at *AT) CatalogEntry() CatalogEntry {
	commentWithStatus := at.commentWithStatus()
	return CatalogEntry{
		Title:  commentWithStatus,
		Method: at.method,
		Path:   at.path,
	}
}

func (at *AT) Resp() *http.Response {
	return at.resp
}

func (at *AT) Err() error {
	return at.err
}

// === Private method ===

func (at *AT) makeURL() *AT {
	// 默认值
	scheme := "http"
	host := "localhost"
	port := ":80"

	if at.scheme != "" {
		scheme = at.scheme
	}
	if at.host != "" {
		host = at.host
	}
	if at.port != "" {
		port = at.port
	}

	path := at.path
	query := ""
	rawurl, err := url.Parse(path)
	if err == nil {
		path = rawurl.Path
		query = rawurl.Query().Encode()
	}
	var realHost string
	if strings.Contains(host, ":") {
		realHost = host
	} else {
		realHost = host + port
	}
	at.url = url.URL{
		Scheme:   scheme,
		Host:     realHost,
		Path:     path,
		RawQuery: query,
	}

	return at
}

func (at *AT) run(realDo bool) *AT {
	// 请求链接
	at = at.makeURL()
	u := at.url

	// 参数处理
	var body = new(bytes.Buffer)
	switch at.method {
	case http.MethodGet, http.MethodDelete:
		q := u.Query()
		if at.param != nil {
			params, err := structToMap(at.param)
			if err != nil {
				at.setErr(err)
				return at
			}
			var valueStr string
			for key, value := range params {
				switch v := value.(type) { // 类型断言，既不能用逗号分隔，也不可用fallthrough
				case []int: // 整型数组
					for _, s := range v {
						valueStr = fmt.Sprintf("%v", s)
						q.Add(key, valueStr)
					}
				case []string: // 字符串数组
					for _, s := range v {
						valueStr = fmt.Sprintf("%v", s)
						q.Add(key, valueStr)
					}
				default:
					valueStr = fmt.Sprintf("%v", value)
					q.Add(key, valueStr)
				}
			}
		}
		u.RawQuery = q.Encode()
	case http.MethodPost, http.MethodPut:
		var paramBytes []byte
		var err error
		switch at.paramFormat {
		case "xml":
			paramBytes, err = xml.Marshal(at.param)
			if err != nil {
				at.setErr(err)
				return at
			}
		default:
			paramBytes, err = json.Marshal(at.param)
			if err != nil {
				at.setErr(err)
				return at
			}
		}
		_, err = body.Write(paramBytes)
		if err != nil {
			at.setErr(err)
			return at
		}
	default:
		at.setErr(fmt.Errorf("not support method %s", at.method))
		return at
	}

	// 文件内容
	var fileContentType string
	if at.file != "" {
		f, err := os.OpenFile(at.file, os.O_RDONLY, os.ModePerm)
		if err != nil {
			at.setErr(err)
			return at
		}
		defer f.Close()

		body.Reset()
		bodyWriter := multipart.NewWriter(body)

		// this step is very important
		fileWriter, err := bodyWriter.CreateFormFile("file", at.file)
		if err != nil {
			at.setErr(err)
			return at
		}

		//iocopy
		_, err = io.Copy(fileWriter, f)
		if err != nil {
			at.setErr(err)
			return at
		}

		fileContentType = bodyWriter.FormDataContentType()
		bodyWriter.Close()
	}

	// 复制一份请求body
	reqBody := make([]byte, body.Len())
	copy(reqBody, body.Bytes())
	at.reqBody = reqBody

	if at.debug {
		fmt.Printf("will do request %s %s with body %s\n", at.method, u.String(), body.String())
	}

	// 新建请求
	req, err := http.NewRequest(at.method, u.String(), body)
	if err != nil {
		at.setErr(err)
		return at
	}

	// 设置header
	innerHeader := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
	if at.paramFormat == "xml" {
		innerHeader = map[string]string{
			"Content-Type": "application/xml; charset=utf-8",
		}
	}
	for headerKey, headerValue := range innerHeader {
		req.Header.Set(headerKey, headerValue)
	}
	for k, v := range at.header {
		for _, vv := range v {
			req.Header.Set(k, vv)
		}
	}
	if fileContentType != "" {
		req.Header.Set("Content-Type", fileContentType)
	}

	// 添加cookie, 支持设置多个
	for _, c := range at.cookies {
		req.AddCookie(c)
	}
	at.req = req

	var tlsConfig *tls.Config
	if at.scheme == "https" {
		if at.insecureSkipVerify {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: at.insecureSkipVerify,
			}
		} else {
			caCrt, err := os.ReadFile(at.caCertPath)
			if err != nil {
				at.setErr(err)
				return at
			}

			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(caCrt)

			cliCrt, err := tls.LoadX509KeyPair(at.certFile, at.keyFile)
			if err != nil {
				at.setErr(err)
				return at
			}

			tlsConfig = &tls.Config{
				RootCAs:      pool,
				Certificates: []tls.Certificate{cliCrt},
			}
		}
	}

	// 发起请求
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxIdleConns:        100, // 最大空闲连接数
		MaxIdleConnsPerHost: 100, // 每个域名最大空闲连接数
		TLSClientConfig:     tlsConfig,
	}
	var clientTimeout = 10 * time.Second
	if at.clientTimeout != 0 {
		clientTimeout = at.clientTimeout
	}
	client := &http.Client{
		Timeout:   clientTimeout, // 超时
		Transport: transport,
	}
	if realDo {
		beforeDo := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			at.setErr(err)
			return at
		}
		afterDo := time.Now()
		used := afterDo.UnixNano() - beforeDo.UnixNano()
		if used >= 1000000000 { // 不小于1s
			if at.isPressureBatch { // 统计数量
				at.slowNum++
			} else {
				fmt.Printf("WARNING: '%s' is slow, used %d ms\n", u.String(), used/1000000)
			}
		}

		// https://stackoverflow.com/questions/17948827/reusing-http-connections-in-golang
		// 只要不关闭response，client就不会重用连接，而是新建连接
		at.resp = resp
	}

	return at
}

const (
	paramName   = "Param"
	returnName  = "Return"
	errorName   = "Error"
	exampleName = "Example"
)

func toAnchor(str string) string {
	return fmt.Sprintf(`<a name="%s" href="#%s">%s</a>`, str, str, str)
}

func (at *AT) commentWithStatus() string {
	commentWithStatus := at.comment
	if at.status != 0 {
		commentWithStatus += " [" + at.status.String() + "]"
	}
	return commentWithStatus
}

var (
	exampleTmpl = `
<details>
<summary>Try to run</summary>
<div>
{{range $k,$v := .Inputs}}<div>
<label for="{{$v.Name}}"><a href="{{$v.Login}}">{{$v.Name}}</a></label>
<p></p>
<textarea rows="4" cols="50" name="{{$v.Name}}" id="{{$v.Id}}" placeholder='{{$v.Placeholder}}'>{{$v.Placeholder}}</textarea>
</div>
{{end}}<div>
<button onclick="sendRequest('{{.Method}}', '{{.Path}}', '{{.Token}}', '{{.Params}}', '{{.ResultDivId}}')">Try to run</button>
<pre id="{{.ResultDivId}}" style="font-size: large"></pre>
</div>
</div>
</details>
`
)

type Example struct {
	Inputs      []Input
	Method      string
	Path        string
	Token       string
	Params      string
	ResultDivId string
}
type Input struct {
	Name        string
	Login       string
	Id          string
	Placeholder string
}

// 生成文档
func (at *AT) makeDoc() *AT {

	var doc string

	// 保存请求和响应
	path := at.path
	key := apiKey(path, at.method)

	commentWithStatus := at.commentWithStatus()

	// 标题
	// 支持anchor：<a name="推送样本到目标" href="#推送样本到目标">推送样本到目标</a>
	doc += "## " + toAnchor(commentWithStatus) + "\n\n"

	// 方法
	doc += "`" + key + "`\n\n"

	// req header
	h := "Request header:\n"
	auth := false
	for k, v := range at.req.Header {
		if k != "Content-Type" && k != at.authHeaderKey {
			continue
		}
		if k == at.authHeaderKey {
			auth = true
		}
		v1 := ""
		if len(v) > 0 {
			v1 = v[0]
		}
		if k == at.authHeaderKey && at.authHeaderValue != "" {
			v1 = at.authHeaderValue
		}
		h += fmt.Sprintf("- %s: %s\n", k, v1)
	}
	if !auth {
		h += fmt.Sprintf("- %s: %s\n", at.authHeaderKey, at.authHeaderValue)
	}
	doc += h + "\n"

	// resp header
	resph := "Response header:\n"
	if at.resp != nil {
		for k, v := range at.resp.Header {
			if k != "Content-Type" {
				continue
			}
			v1 := ""
			if len(v) > 0 {
				v1 = v[0]
			}
			resph += fmt.Sprintf("- %s: %s\n", k, v1)
		}
		resph += "\n"
	} else {
		switch at.resultFormat {
		case "xml":
			resph += "- Content-Type: application/xml; charset=utf-8\n\n"
		default:
			resph += "- Content-Type: application/json; charset=utf-8\n\n"
		}
	}
	doc += resph

	// 在解析参数和返回的同时，收集注释信息：map[string]string, 其中key的值需要保留每层的路径，如：|list|name
	// 参数
	block, pkcm, err := structToBlock(paramName, at.method, at.param)
	if err != nil {
		at.setErr(err)
		return at
	}
	doc += block

	// 返回
	block, rkcm, err := structToBlock(returnName, at.method, at.result)
	if err != nil {
		at.setErr(err)
		return at
	}
	doc += block

	// 错误码
	if len(at.ates) > 0 {
		block, err = structToList(errorName, at.ates...)
		if err != nil {
			at.setErr(err)
			return at
		}
		doc += block
	}

	var paramData []byte
	switch at.method {
	case http.MethodGet, http.MethodDelete:
		paramData = []byte(at.req.URL.RawQuery)
	case http.MethodPost, http.MethodPut:
		paramData = at.reqBody
	}

	paramId := "param" + at.path + " " + at.method
	tokenId := "token" + at.path + " " + at.method
	resultDivId := "result" + at.path + " " + at.method
	var inputs []Input
	if len(pkcm) > 0 {
		inputs = append(inputs, Input{Name: "Params(参照下面的示例)", Id: paramId, Placeholder: string(paramData)})
	}
	if at.authHeaderKey != "" {
		inputs = append(inputs, Input{Name: "Token(从登录接口返回 - " + http.CanonicalHeaderKey(at.authHeaderKey) + ": " + at.authHeaderValue + ")", Login: "/apidoc/README", Id: tokenId, Placeholder: "TOKEN STRING"})
	}
	buf := new(bytes.Buffer)
	do.Must1(template.New(resultDivId).Parse(exampleTmpl)).Execute(buf, Example{
		Inputs:      inputs,
		Method:      strings.ToLower(at.method),
		Path:        at.path,
		Token:       tokenId,
		Params:      paramId,
		ResultDivId: resultDivId,
	})
	doc += buf.String() + "\n\n"

	doc += exampleName + ":\n\n"

	// 参数和返回示例
	switch at.method {
	case http.MethodGet, http.MethodDelete:
		doc += dataToSummary(paramName, []byte(at.req.URL.RawQuery), at.paramFormat, false, nil)
	case http.MethodPost, http.MethodPut:
		isjson := at.file == ""
		doc += dataToSummary(paramName, at.reqBody, at.paramFormat, isjson, pkcm)
	}

	// 复制resp.Body
	var data []byte
	if at.resp != nil {
		data, _, err = copyResponseBody(at.resp)
		if err != nil {
			at.setErr(err)
			return at
		}
	} else {
		switch at.resultFormat {
		case "xml":
			data, err = xml.Marshal(at.result)
			if err != nil {
				at.setErr(err)
				return at
			}
		default:
			data, err = json.Marshal(at.result)
			if err != nil {
				at.setErr(err)
				return at
			}
		}
	}
	doc += dataToSummary(returnName, data, at.resultFormat, true, rkcm)

	at.doc = doc

	return at
}

func (at *AT) setErr(err error) *AT {
	if at.err == nil {
		at.err = err
	}
	return at
}

func (at *AT) jsonIndent(w io.Writer, r any) *AT {
	if at.debug {
		JSONIndent(w, r)
	}
	return at
}

func (at *AT) clone() *AT {
	return NewAT(at.path, at.method, at.comment, at.header, at.cookies)
}

var (
	_ = (&AT{}).registerHandler("http.Handler", http.Handler(nil))
	_ = (&AT{}).registerHandler("http.HandlerFunc", http.HandlerFunc(nil))
)

// 提供`registerHandler`方法，传入的handler参数值，应该是(http.Handler)(nil)
func (at *AT) registerHandler(name string, handler any) *AT {
	if at.handlerMap == nil {
		at.handlerMap = make(map[string]any)
	}
	// 解析handler，得到其接口类型，后续解析源码时，寻找该类型的变量，进而生成代码
	// 如果可以的话，在寻找到handler之后，在handler里解析出param和result结构体，
	// 从而得到接口定义里最需要的三个信息：路由、参数、返回。
	at.handlerMap[name] = handler

	return at
}

type (
	DocHelper interface {
		Fatal(args ...any)
		FindTestAPIsByPrefix(prefix string) (r []*TestAPI)
		GetParamResult(key string, param reflect.Type, result reflect.Type) (p any, r any)
	}
)

func MakeDoc(t DocHelper, dir, file, title, pathPrefix string) {
	pf, err := OpenFile(filepath.Join(dir, file), title)
	if err != nil {
		t.Fatal(err)
	}
	f := new(bytes.Buffer)
	catalogs := []CatalogEntry{}
	defer func() {
		catalog, err := MakeCatalog(catalogs)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pf.Write([]byte(catalog)); err != nil {
			t.Fatal(err)
		}
		if _, err := pf.Write(f.Bytes()); err != nil {
			t.Fatal(err)
		}
		pf.Close()
	}()

	// doc
	for _, item := range t.FindTestAPIsByPrefix(pathPrefix) {
		{
			at := item
			p, r := at.GetParamResult(t.GetParamResult)
			if err = at.SetParam(p).
				FakeRun().
				Result(r).
				Errors().
				WriteFile(f).
				Err(); err != nil {
				t.Fatal(err)
			}
			catalogs = append(catalogs, at.CatalogEntry())
		}
	}
}
