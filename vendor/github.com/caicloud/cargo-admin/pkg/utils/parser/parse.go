package parser

import (
	"fmt"
	"strconv"
	"strings"
)

func HostPort(addr string) (host string, port uint16, err error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("address: %s is invalid", addr)
	}

	port64, err := strconv.ParseUint(parts[1], 10, 16)
	if err != nil {
		return "", 0, err
	}

	return parts[0], uint16(port64), nil
}

type Image struct {
	Project string
	Repo    string
	Tag     string
}

func ImageName(name string) (*Image, error) {
	repo := strings.SplitN(name, "/", 2)
	if len(repo) < 2 {
		return nil, fmt.Errorf("unable to parse image from string: %s", name)
	}
	i := strings.SplitN(repo[1], ":", 2)
	res := &Image{
		Project: repo[0],
		Repo:    i[0],
	}
	if len(i) == 2 {
		res.Tag = i[1]
	}
	return res, nil
}
