package gock

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/nbio/st"
)

func TestTransportMatch(t *testing.T) {
	defer after()
	const uri = "http://foo.com"
	New(uri).Reply(204)
	u, _ := url.Parse(uri)
	req := &http.Request{URL: u}
	res, err := NewTransport().RoundTrip(req)
	st.Expect(t, err, nil)
	st.Expect(t, res.StatusCode, 204)
	st.Expect(t, res.Request, req)
}

func TestTransportCannotMatch(t *testing.T) {
	defer after()
	New("http://foo.com").Reply(204)
	u, _ := url.Parse("http://127.0.0.1:1234")
	req := &http.Request{URL: u}
	_, err := NewTransport().RoundTrip(req)
	st.Expect(t, err, ErrCannotMatch)
}

func TestTransportNotIntercepting(t *testing.T) {
	defer after()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world")
	}))
	defer ts.Close()

	New(ts.URL).Reply(200)
	Disable()

	u, _ := url.Parse(ts.URL)
	req := &http.Request{URL: u, Header: make(http.Header)}

	res, err := NewTransport().RoundTrip(req)
	st.Expect(t, err, nil)
	st.Expect(t, Intercepting(), false)
	st.Expect(t, res.StatusCode, 200)
}
