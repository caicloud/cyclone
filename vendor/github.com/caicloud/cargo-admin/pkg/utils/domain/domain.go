package domain

import (
	"net/url"

	. "github.com/caicloud/cargo-admin/pkg/errors"

	"github.com/caicloud/nirvana/log"
)

func GetDomain(host string) (string, error) {
	url, err := url.Parse(host)
	if err != nil {
		log.Errorf("url.Parse err: %v, host: %s", err, host)
		return "", ErrorUnknownInternal.Error(err)
	}
	return url.Host, nil
}
