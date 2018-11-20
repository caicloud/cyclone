package testtools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"

	constants "github.com/caicloud/storage-admin/pkg/constants"
)

type Url struct {
	host     string
	port     int
	rootPath string
	subPath  string
	paramMap map[string]string
}

func NewUrl() *Url {
	return new(Url)
}

func (u *Url) Host(host string) *Url {
	if strings.HasPrefix(host, "http://") {
		u.host = host[7:]
	} else {
		u.host = host
	}
	return u
}
func (u *Url) Port(port int) *Url {
	u.port = port
	return u
}
func (u *Url) RootPath(rootPath string) *Url {
	u.rootPath = rootPath
	return u
}
func (u *Url) RootPathDefault() *Url {
	u.rootPath = constants.RootPath
	return u
}
func (u *Url) SubPath(subPath string) *Url {
	u.subPath = subPath
	return u
}
func (u *Url) Param(k, v string) *Url {
	if u.paramMap == nil {
		u.paramMap = map[string]string{k: v}
	}
	u.paramMap[k] = v
	return u
}
func (u *Url) String() string {
	hostPort := u.host
	if u.port > 0 {
		hostPort = fmt.Sprintf("%s:%d", u.host, u.port)
	}
	re := path.Join(hostPort, u.rootPath, u.subPath)
	return "http://" + re + ParameterString(u.paramMap)
}

func ParameterString(pm map[string]string) string {
	if len(pm) == 0 {
		return ""
	}
	ss := make([]string, 0, len(pm))
	for k, v := range pm {
		k = url.QueryEscape(k)
		v = url.QueryEscape(v)
		ss = append(ss, strings.Join([]string{k, v}, "="))
	}
	sort.Strings(ss)
	return "?" + strings.Join(ss, "&")
}

func SetPathParam(base, name, value string) string {
	value = url.PathEscape(value)
	return strings.Replace(base, fmt.Sprintf("{%s}", name), value, -1)
}

func GetData(u *Url) (int, []byte, error) {
	req, e := http.NewRequest(http.MethodGet, u.String(), nil)
	if e != nil {
		return 0, nil, e
	}
	return doRequest(req)
}
func PostData(u *Url, data []byte) (int, []byte, error) {
	req, e := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(data))
	if e != nil {
		return 0, nil, e
	}
	return doRequest(req)
}
func PutData(u *Url, data []byte) (int, []byte, error) {
	req, e := http.NewRequest(http.MethodPut, u.String(), bytes.NewReader(data))
	if e != nil {
		return 0, nil, e
	}
	return doRequest(req)
}

func Get(u *Url, okCode int, resp interface{}) (code int, b []byte, e error) {
	code, b, e = GetData(u)
	if e != nil {
		return
	}
	e = ParseResponse(code, b, okCode, resp)
	return
}
func Post(u *Url, okCode int, req, resp interface{}) (code int, b []byte, e error) {
	var data []byte
	data, e = json.Marshal(req)
	if e != nil {
		return
	}
	code, b, e = PostData(u, data)
	if e != nil {
		return
	}
	e = ParseResponse(code, b, okCode, resp)
	return
}
func Put(u *Url, okCode int, req, resp interface{}) (code int, b []byte, e error) {
	var data []byte
	data, e = json.Marshal(req)
	if e != nil {
		return
	}
	code, b, e = PutData(u, data)
	if e != nil {
		return
	}
	e = ParseResponse(code, b, okCode, resp)
	return
}
func Delete(u *Url, okCode int) (code int, b []byte, e error) {
	var req *http.Request
	req, e = http.NewRequest(http.MethodDelete, u.String(), nil)
	if e != nil {
		return
	}
	code, b, e = doRequest(req)
	if e != nil {
		return
	}
	e = ParseResponse(code, b, okCode, nil)
	return
}

func doRequest(req *http.Request) (int, []byte, error) {
	req.Header.Set("Content-Type", constants.MimeJson)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return 0, nil, e
	}
	defer resp.Body.Close()
	b, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return resp.StatusCode, nil, e
	}
	return resp.StatusCode, b, nil
}

func ParseResponse(code int, data []byte, expectedCode int, expectedRet interface{}) error {
	if code != expectedCode {
		return fmt.Errorf("unexpected http code [%d!=%d], err %v", code, expectedCode, string(data))
	}
	if expectedRet == nil {
		return nil
	}
	e := json.Unmarshal(data, expectedRet)
	if e != nil {
		return e
	}
	cb, e := json.Marshal(expectedRet)
	if e != nil {
		return e
	}
	if isListResponse(cb) && !isListResponse(data) {
		return fmt.Errorf("list object parse failed")
	}
	if isObjectResponse(cb) && !isObjectResponse(data) {
		return fmt.Errorf("single object parse failed")
	}
	return nil
}

func isListResponse(b []byte) bool {
	return bytes.Contains(b, []byte("\"metadata\":")) &&
		bytes.Contains(b, []byte("\"total\":")) &&
		bytes.Contains(b, []byte("\"items\":"))
}

func isObjectResponse(b []byte) bool {
	return !isListResponse(b) &&
		bytes.Contains(b, []byte("\"metadata\":")) &&
		(bytes.Contains(b, []byte("\"creationTime\":")) || bytes.Contains(b, []byte("\"creationTimestamp\":")))
}
