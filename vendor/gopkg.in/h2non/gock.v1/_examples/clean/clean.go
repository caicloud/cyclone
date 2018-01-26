package main

import (
	"fmt"
	"gopkg.in/h2non/gock.v1"
	"net/http"
)

func main() {
	defer gock.Disable()
	defer gock.Flush() // Flush all the registered mocks, including the pending ones.
	defer gock.Clean() // Clean all the finished mocks, but keeping the pending ones.
	// defer gock.Off() -> Or you can simply call Off() method

	gock.New("http://httpbin.org").
		Get("/get").
		Filter(func(req *http.Request) bool { return req.URL.Host == "httpbin.org" }).
		Filter(func(req *http.Request) bool { return req.URL.Path == "/get" }).
		Reply(204).
		SetHeader("Server", "gock")

	res, err := http.Get("http://httpbin.org/get")
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Server header: %s\n", res.Header.Get("Server"))
}
