package httpclient

import (
    "bytes"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "strconv"
    "sync"

    jsoniter "github.com/json-iterator/go"
    "github.com/pkg/errors"
)

type requestMethod = string
type contentType = string

const (
    reqOption  requestMethod = "OPTIONS"
    reqGet     requestMethod = "GET"
    reqHead    requestMethod = "HEAD"
    reqPost    requestMethod = "POST"
    reqPut     requestMethod = "PUT"
    reqDelete  requestMethod = "DELETE"
    reqTrace   requestMethod = "TRACE"
    reqConnect requestMethod = "CONNECT"

    ctypeURLEncoded contentType = "application/x-www-form-urlencoded"
    ctypeJson       contentType = "application/json"
)

type Client struct {
    *http.Client
}

var DefaultHTTPClient = &Client{http.DefaultClient}

// request
type request struct {
    err     error
    client  *Client
    request *http.Request

    query url.Values
}

// newRequest
func (c *Client) newRequest(rm requestMethod, u string) *request {
    ret := &request{
        client: c,
    }
    req, err := http.NewRequest(rm, u, nil)
    ret.request = req
    ret.err = err
    ret.query = url.Values{}
    return ret
}

// AddQuery
func (c *request) AddQuery(key, value string) *request {
    c.query.Add(key, value)
    return c
}

// AddQuery4Json
// 将对象转化为请求行参数
func (c *request) AddQuery4Json(obj interface{}) *request {
    data, err := jsoniter.Marshal(obj)
    if err != nil && c.err == nil {
        c.err = err
    }
    m := make(map[string]interface{})
    err = jsoniter.Unmarshal(data, m)
    if err != nil && c.err == nil {
        c.err = err
    }
    for k, v := range m {
        c.AddQuery(k, fmt.Sprintf("%v", v))
    }
    return c
}

// SetHeader SetHeader
func (c *request) SetHeader(key, value string) *request {
    c.Header().Set(key, value)
    return c
}

// AddHeader AddHeader
func (c *request) AddHeader(key, value string) *request {
    c.Header().Add(key, value)
    return c
}

// SetCookie SetCookie
func (c *request) SetCookie(cookie *http.Cookie) *request {
    c.request.AddCookie(cookie)
    return c
}

// Body Body
func (c *request) SetBody(ct contentType, length int, body io.ReadCloser) *request {
    c.request.Body = body
    c.SetHeader("Content-Type", ct)
    c.SetHeader("Content-Length", strconv.Itoa(length))
    return c
}

// SetBody4Bytes
func (c *request) SetBody4Bytes(ct contentType, data []byte) *request {
    buf := io.NopCloser(bytes.NewBuffer(data))
    return c.SetBody(ct, len(data), buf)
}

// SetBody4Values
// 参数写入消息体
func (c *request) SetBody4Values(v url.Values) *request {
    return c.SetBody4Bytes(ctypeURLEncoded, []byte(v.Encode()))
}

// SetBody4Json
// 参数写入消息体
func (c *request) SetBody4Json(obj interface{}) *request {
    data, err := jsoniter.Marshal(obj)
    if err != nil && c.err == nil {
        c.err = err
    }
    return c.SetBody4Bytes(ctypeJson, data)
}

// Header Header
func (c *request) Header() (header http.Header) {
    return c.request.Header
}

// Cookie Cookie
func (c *request) Cookie() []*http.Cookie {
    return c.request.Cookies()
}

// Body Body
func (c *request) Body() io.ReadCloser {
    return c.request.Body
}

// Do Do
func (c *request) Do() *Response {
    if c.err != nil {
        return newResponse(nil, c.err)
    }
    c.request.URL.RawQuery += "&" + c.query.Encode()
    resp, err := c.client.Do(c.request)
    return newResponse(resp, err)
}

func Get(url string) *request {
    return DefaultHTTPClient.newRequest(reqGet, url)
}

func (c *Client) Get(url string) *request {
    return c.newRequest(reqGet, url)
}

func Post(url string) *request {
    return DefaultHTTPClient.newRequest(reqPost, url)
}

// Post
func (c *Client) Post(url string) *request {
    return c.newRequest(reqPost, url)
}

// Response
type Response struct {
    // resp 有可能为空
    resp *http.Response
    err  error
    data []byte
    o    sync.Once
}

func newResponse(resp *http.Response, err error) *Response {
    return &Response{
        resp: resp,
        err:  err,
    }
}

// Error Error
func (r *Response) Error() error {
    return r.err
}

// Response 将响应交给调用控制
func (r *Response) Response() *http.Response {
    return r.resp
}

// Header Header
func (r *Response) Header() (header http.Header) {
    if r.resp == nil {
        return http.Header{}
    }
    return r.resp.Header
}

// Status
func (r *Response) Status() string {
    if r.resp == nil {
        return ""
    }
    return r.resp.Status
}

// StatusCode
func (r *Response) StatusCode() int {
    if r.resp == nil {
        return 0
    }
    return r.resp.StatusCode
}

// ToBytes
func (r *Response) ToBytes() (data []byte, err error) {
    if r.data != nil {
        // 已经取过数据 直接返回
        return r.data, r.err
    }
    if r.resp != nil {
        // 读取响应体
        r.o.Do(func() {
            defer r.resp.Body.Close()
            r.data, err = ioutil.ReadAll(r.resp.Body)
            if r.err == nil {
                r.err = err
            }
        })
    }
    if r.err != nil {
        return nil, r.err
    }
    if r.StatusCode() != 200 {
        // 封装应用错误
        buf := bytes.NewBuffer(nil)
        buf.WriteString(r.Status())
        buf.WriteByte(' ')
        buf.Write(r.data)
        r.err = errors.New(buf.String())
    }
    return r.data, r.err
}

// ToJson
func (r *Response) ToJson(obj interface{}) (err error) {
    msg, err := r.ToBytes()
    if err != nil {
        return err
    }
    return jsoniter.Unmarshal(msg, obj)
}

func (r *Response) ToString() (string, error) {
    msg, err := r.ToBytes()
    return string(msg), err
}
