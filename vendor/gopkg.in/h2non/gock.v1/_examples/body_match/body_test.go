package test

import (
	"bytes"
	"github.com/nbio/st"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestMockSimple(t *testing.T) {
	defer gock.Off()

	gock.New("http://foo.com").
		Post("/bar").
		MatchType("json").
		JSON(map[string]string{"foo": "bar"}).
		Reply(201).
		JSON(map[string]string{"bar": "foo"})

	body := bytes.NewBuffer([]byte(`{"foo":"bar"}`))
	res, err := http.Post("http://foo.com/bar", "application/json", body)
	st.Expect(t, err, nil)
	st.Expect(t, res.StatusCode, 201)

	resBody, _ := ioutil.ReadAll(res.Body)
	st.Expect(t, string(resBody)[:13], `{"bar":"foo"}`)
}
