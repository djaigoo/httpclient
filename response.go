package httpclient

import (
    "encoding/json"
    "io/ioutil"
    "net/http"
)

// Response
type Response struct {
    resp *http.Response
    err  error
}

func newResponse(resp *http.Response, err error) Response {
    return Response{
        resp: resp,
        err:  err,
    }
}

func (r Response) Error() error {
    return r.err
}

func (r Response) Response() *http.Response {
    return r.resp
}

func (r Response) Headers() (header http.Header) {
    if r.resp == nil {
        return http.Header{}
    }
    return r.resp.Header
}

func (r Response) GetCookie(name string) *http.Cookie {
    for _, cookie := range r.Cookies() {
        if cookie.Name == name {
            return cookie
        }
    }
    return nil
}

func (r Response) Cookies() []*http.Cookie {
    return r.resp.Cookies()
}

func (r Response) ToText() (msg []byte, err error) {
    if r.resp != nil {
        defer r.resp.Body.Close()
    }
    if r.err != nil {
        return nil, r.err
    }
    if r.resp.Body == nil {
        return nil, nil
    }
    return ioutil.ReadAll(r.resp.Body)
}

func (r Response) ToJson(obj interface{}) (err error) {
    msg, err := r.ToText()
    if err != nil {
        return err
    }
    return json.Unmarshal(msg, obj)
}

func (r Response) ToString() (string, error) {
    msg, err := r.ToText()
    return string(msg), err
}
