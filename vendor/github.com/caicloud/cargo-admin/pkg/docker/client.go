package docker

import (
	"log"

	"github.com/docker/docker/client"
)

var Client *client.Client

func init() {
	c, err := client.NewClient("unix:///var/run/dind/docker.sock", "", nil, nil)
	if err != nil {
		log.Fatal("Create docker client failed")
	}
	Client = c
}
