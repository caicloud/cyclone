package main

import (
	"fmt"
	"gopkg.in/h2non/gock.v1"
	"net/http"
)

func main() {
	defer gock.Disable()

	gock.New("http://httpbin.org").
		Get("/").
		Map(func(req *http.Request) *http.Request { req.URL.Host = "httpbin.org"; return req }).
		Map(func(req *http.Request) *http.Request { req.URL.Path = "/"; return req }).
		Reply(204).
		SetHeader("Server", "gock")

	res, err := http.Get("http://httpbin.org/get")
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Server header: %s\n", res.Header.Get("Server"))
}
