package test

import (
	"github.com/nbio/st"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestMatchQueryParams(t *testing.T) {
	defer gock.Disable()

	gock.New("http://foo.com").
		MatchParam("foo", "^bar$").
		MatchParam("bar", "baz").
		ParamPresent("baz").
		Reply(200).
		BodyString("foo foo")

	req, err := http.NewRequest("GET", "http://foo.com?foo=bar&bar=baz&baz=foo", nil)
	res, err := (&http.Client{}).Do(req)
	st.Expect(t, err, nil)
	st.Expect(t, res.StatusCode, 200)
	body, _ := ioutil.ReadAll(res.Body)
	st.Expect(t, string(body), "foo foo")
}
