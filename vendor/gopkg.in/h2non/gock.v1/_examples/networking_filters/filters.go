package main

import (
	"fmt"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
)

func main() {
	defer gock.Disable()
	defer gock.DisableNetworking()
	defer gock.DisableNetworkingFilters()

	gock.EnableNetworking()

	// Register a networking filter
	gock.NetworkingFilter(func(req *http.Request) bool {
		return req.URL.Host == "httpbin.org"
	})

	gock.New("http://httpbin.org").
		Get("/get").
		Reply(201).
		SetHeader("Server", "gock")

	res, err := http.Get("http://httpbin.org/get")
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}

	// The response status comes from the mock
	fmt.Printf("Status: %d\n", res.StatusCode)
	// The server header comes from mock as well
	fmt.Printf("Server header: %s\n", res.Header.Get("Server"))
	// Response body is the original
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Printf("Body: %s", string(body))
}
