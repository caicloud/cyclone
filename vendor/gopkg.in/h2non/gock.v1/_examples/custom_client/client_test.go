package test

import (
	"github.com/nbio/st"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestClient(t *testing.T) {
	defer gock.Off()

	gock.New("http://foo.com").
		Reply(200).
		BodyString("foo foo")

	req, err := http.NewRequest("GET", "http://foo.com", nil)
	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)

	res, err := client.Do(req)
	st.Expect(t, err, nil)
	st.Expect(t, res.StatusCode, 200)
	body, _ := ioutil.ReadAll(res.Body)
	st.Expect(t, string(body), "foo foo")
}
