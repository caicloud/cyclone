package test

import (
	"github.com/nbio/st"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestSimple(t *testing.T) {
	defer gock.Off()

	gock.New("http://foo.com").
		Get("/bar").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	res, err := http.Get("http://foo.com/bar")
	st.Expect(t, err, nil)
	st.Expect(t, res.StatusCode, 200)

	body, _ := ioutil.ReadAll(res.Body)
	st.Expect(t, string(body)[:13], `{"foo":"bar"}`)

	// Verify that we don't have pending mocks
	st.Expect(t, gock.IsDone(), true)
}
