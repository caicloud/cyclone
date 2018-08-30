package parser

import (
	"reflect"
	"testing"
)

func TestHostPort(t *testing.T) {
	cases := []struct {
		a string
		h string
		p uint16
	}{
		{
			"localhost:8080",
			"localhost",
			8080,
		},
		{
			"0.0.0.0:8081",
			"0.0.0.0",
			8081,
		},
	}

	for _, c := range cases {
		h, p, _ := HostPort(c.a)
		if h != c.h || p != c.p {
			t.Errorf("parse host port error, expected host=%s, port=%d, but got host=%s, port=%d", c.h, c.p, h, p)
		}
	}
}

func TestImageName(t *testing.T) {
	cases := []struct {
		n string
		i *Image
	}{
		{
			"library/busybox:latest",
			&Image{"library", "busybox", "latest"},
		},
		{
			"library/busybox",
			&Image{"library", "busybox", ""},
		},
		{
			"busybox",
			nil,
		},
	}

	for _, c := range cases {
		i, _ := ImageName(c.n)
		if !reflect.DeepEqual(i, c.i) {
			t.Errorf("parse image name error, expected %v, but got %v", c.i, i)
		}
	}
}
