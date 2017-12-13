package cyclone_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"time"
)

type testCase struct {
	testReq *req
	testRsp *rsp
}

type header struct {
	k string
	v string
}

type req struct {
	uri    string
	method string
	body   interface{}
}

type rsp struct {
	statusCode int
	result     interface{}
	check      interface{}
}

type testCases struct {
	cases   []*testCase
	headers []*header
}

type TestCases interface {
	Add(string, string, interface{}, interface{}, interface{}, int)
	Test()
	Clear()
}

func NewTestCases(head []*header) TestCases {
	cases := new(testCases)
	cases.cases = nil
	cases.headers = head

	return cases
}

// Add adds testCase with req(uri, method, reqBody),
func (t *testCases) Add(uri string, method string, reqBody, rspCheck, rspResult interface{}, statusCode int) {
	testReq := &req{
		uri:    uri,
		method: method,
		body:   reqBody,
	}

	testRsp := &rsp{
		statusCode: statusCode,
		check:      rspCheck,
		result:     rspResult,
	}

	t.cases = append(t.cases, &testCase{
		testReq: testReq,
		testRsp: testRsp,
	})
}

func (t *testCases) Clear() {
	t.cases = nil
	t.headers = nil
}

func (t *testCases) Test() {
	for _, testCase := range t.cases {
		re := testCase.testReq
		rs := testCase.testRsp
		code, err := restApiCall(re.uri, re.method, re.body, rs.result, t.headers)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).Should(Equal(rs.statusCode))
		if rs.check != nil {
			Expect(rs.result).Should(Equal(rs.check))
		}
	}
}

func restApiCall(uri string, method string, req, rsp interface{}, headers []*header) (int, error) {
	transport := &http.Transport{
		DisableKeepAlives: true,
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1%s", port, uri)

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second,
	}

	reqBody, err := json.Marshal(req)
	Expect(err).ShouldNot(HaveOccurred())

	request, errReq := http.NewRequest(method, url, bytes.NewReader(reqBody))
	Expect(errReq).ShouldNot(HaveOccurred())

	for _, header := range headers {
		request.Header.Set(header.k, header.v)
	}

	response, errCall := client.Do(request)
	Expect(errCall).ShouldNot(HaveOccurred())
	if rsp != nil {
		body, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(body, rsp)
	}

	return response.StatusCode, errCall
}
