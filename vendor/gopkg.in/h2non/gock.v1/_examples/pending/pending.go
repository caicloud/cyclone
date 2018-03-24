package main

import (
	"fmt"
	"gopkg.in/h2non/gock.v1"
	"net/http"
)

func main() {
	defer gock.Off()

	gock.New("http://httpbin.org").
		Get("/get").
		Reply(204).
		SetHeader("Server", "gock")

	fmt.Printf("Pending mocks before request: %d\n", len(gock.Pending()))
	fmt.Printf("Is pending before request: %#v\n", gock.IsPending())

	_, err := http.Get("http://httpbin.org/get")
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}

	fmt.Printf("Pending mocks after request: %d\n", len(gock.Pending()))
	fmt.Printf("Is pending: %#v\n", gock.IsPending())
}
