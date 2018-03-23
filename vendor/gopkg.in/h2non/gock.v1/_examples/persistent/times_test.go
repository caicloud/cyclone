package test

import (
	"github.com/nbio/st"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestTimes(t *testing.T) {
	defer gock.Disable()
	gock.New("http://127.0.0.1:1234").
		Get("/bar").
		Times(4).
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	for i := 0; i < 5; i++ {
		res, err := http.Get("http://127.0.0.1:1234/bar")
		if i == 4 {
			st.Reject(t, err, nil)
			break
		}

		st.Expect(t, err, nil)
		st.Expect(t, res.StatusCode, 200)
		body, _ := ioutil.ReadAll(res.Body)
		st.Expect(t, string(body)[:13], `{"foo":"bar"}`)
	}
}
