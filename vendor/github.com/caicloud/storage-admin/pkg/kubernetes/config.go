package kubernetes

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/rest"
)

func ParseConfigFromRequest(req *http.Request) (*Config, error) {
	host := req.Header.Get(HeaderK8SHost)
	if len(host) == 0 {
		return nil, fmt.Errorf("get %s from request header failed", HeaderK8SHost)
	}

	if token := req.Header.Get(HeaderToken); len(token) > 0 {
		return &Config{
			Host:        host,
			BearerToken: token,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}, nil
	}

	user := req.Header.Get(HeaderUsername)
	if len(user) == 0 {
		return nil, fmt.Errorf("bad %s", HeaderUsername)
	}
	pwd := req.Header.Get(HeaderPassword)
	if len(pwd) == 0 {
		return nil, fmt.Errorf("bad %s", HeaderPassword)
	}
	return &Config{
		Host:     host,
		Username: user,
		Password: pwd,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}, nil
}

func IsConfigSame(a, b *Config) bool {
	if a == b {
		return true
	}
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}
	if a.Host != b.Host {
		return false
	}
	if a.BearerToken != b.BearerToken {
		return false
	}
	if len(a.BearerToken) == 0 {
		return a.Username == b.Username && a.Password == b.Password
	}
	return true
}
