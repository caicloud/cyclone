package cauth

import "net/http"

type CauthClient struct {
	Client   *http.Client
	Host     string
	Registry string
}

func NewClient(cauth, registry string) *CauthClient {
	return &CauthClient{
		Client:   http.DefaultClient,
		Host:     cauth,
		Registry: registry,
	}
}
