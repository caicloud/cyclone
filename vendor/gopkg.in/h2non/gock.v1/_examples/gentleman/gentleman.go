package main

import (
	"fmt"

	"gopkg.in/h2non/gentleman.v1"
	"gopkg.in/h2non/gentleman.v1/context"
	"gopkg.in/h2non/gock.v1"
)

// Usege example with gentleman HTTP client toolkit.
// See also: https://github.com/h2non/gentleman-mock
func main() {
	defer gock.Off()

	gock.New("http://httpbin.org").
		Get("/*").
		Reply(204).
		SetHeader("Server", "gock")

	cli := gentleman.New()

	cli.UseHandler("before dial", func(ctx *context.Context, h context.Handler) {
		gock.InterceptClient(ctx.Client)
		h.Next(ctx)
	})

	res, err := cli.Request().URL("http://httpbin.org/get").Send()
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Server header: %s\n", res.Header.Get("Server"))
}
