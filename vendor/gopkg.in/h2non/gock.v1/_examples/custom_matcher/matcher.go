package main

import (
	"fmt"
	"gopkg.in/h2non/gock.v1"
	"net/http"
)

func main() {
	defer gock.Off()

	// Create a new custom matcher with HTTP headers only matchers
	matcher := gock.NewBasicMatcher()

	// Add a custom match function
	matcher.Add(func(req *http.Request, ereq *gock.Request) (bool, error) {
		return req.URL.Scheme == "http", nil
	})

	// Define the mock
	gock.New("http://httpbin.org").
		SetMatcher(matcher).
		Get("/").
		Reply(204).
		SetHeader("Server", "gock")

	res, err := http.Get("http://httpbin.org/get")
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Server header: %s\n", res.Header.Get("Server"))
}
