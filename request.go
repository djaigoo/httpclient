package http

import (
    "bytes"
    "encoding/json"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "strconv"
)

const (
    optionsReq = "OPTIONS"
    getReq     = "GET"
    headReq    = "HEAD"
    postReq    = "POST"
    putReq     = "PUT"
    deleteReq  = "DELETE"
    traceReq   = "TRACE"
    connectReq = "CONNECT"
    
    contentType   = "Content-Type"
    contentLength = "Content-Length"
)

// client
type client struct {
    err     error
    client  *http.Client
    request *http.Request
}

func newClient(method string, url string) (c client) {
    ret := client{
        client: http.DefaultClient,
    }
    request, err := http.NewRequest(method, url, nil)
    ret.request = request
    ret.err = err
    return ret
}

func (c client) SetHeader(key, value string) client {
    c.Headers().Set(key, value)
    return c
}

func (c client) AddHeader(key, value string) client {
    c.Headers().Add(key, value)
    return c
}

func (c client) DelHeader(key string) client {
    c.Headers().Del(key)
    return c
}

func (c client) Headers() (header http.Header) {
    return c.request.Header
}

func (c client) SetCookie(cookie *http.Cookie) client {
    c.request.AddCookie(cookie)
    return c
}

func (c client) SetBody(body io.Reader) client {
    rc, ok := body.(io.ReadCloser)
    if !ok && body != nil {
        rc = ioutil.NopCloser(body)
    }
    c.request.Body = rc
    return c
}

func (c client) Cookie() []*http.Cookie {
    return c.request.Cookies()
}

func (c client) Body() io.ReadCloser {
    return c.request.Body
}

func (c client) Do() Response {
    if c.err != nil {
        return Response{nil, c.err}
    }
    resp, err := c.client.Do(c.request)
    return newResponse(resp, err)
}

type getEntity struct {
    client
}

func Get(url string) getEntity {
    return getEntity{client: newClient(getReq, url)}
}

func GetAsJson(url string, obj interface{}) (err error) {
    return Get(url).Do().ToJson(obj)
}

type postEntity struct {
    client
}

func (p postEntity) ValueBody(v url.Values) Response {
    return p.TextBody([]byte(v.Encode()))
}

func (p postEntity) TextBody(data []byte) Response {
    return p.Do("application/x-www-form-urlencoded", data)
}

func (p postEntity) JsonBody(obj interface{}) Response {
    data, err := json.Marshal(obj)
    if err != nil {
        return newResponse(nil, err)
    }
    return p.Do("application/json", data)
}

func (p postEntity) Do(ctype string, data []byte) Response {
    body := bytes.NewBuffer(data)
    rc := ioutil.NopCloser(body)
    p.SetHeader(contentLength, strconv.Itoa(len(data))).SetHeader(contentType, ctype).SetBody(rc)
    return p.client.Do()
}

func Post(url string) postEntity {
    return postEntity{client: newClient(postReq, url)}
}
