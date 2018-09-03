package resource

import (
	"net"
)

func retry(err error) bool {
	if err == nil {
		return false
	}
	return isNetworkErr(err)
}

func isTemporary(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary()
	}
	return false
}

func isNetworkErr(err error) bool {
	_, ok := err.(net.Error)
	return ok
}
