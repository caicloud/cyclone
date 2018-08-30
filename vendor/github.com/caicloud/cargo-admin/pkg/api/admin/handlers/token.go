package handlers

import (
	"context"

	"github.com/caicloud/cargo-admin/pkg/api/token/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"

	"github.com/caicloud/cargo-admin/pkg/token"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
)

func GetToken(ctx context.Context) (*types.Token, error) {
	req := service.HTTPContextFrom(ctx).Request()
	scope := req.URL.Query()["scope"]
	host := req.Host
	if host == "" {
		log.Error("Host in request is empay")
		return nil, ErrorUnknownRequest.Error("Host field must be set in request")
	}
	log.Infof("request host: %s", host)

	if len(req.Header["Authorization"]) == 0 {
		log.Infof("anonymous access")
		return token.Get(host, "", "", scope)
	}

	user, pwd, ok := req.BasicAuth()
	if !ok {
		return nil, ErrorUnauthentication.Error()
	}
	log.Infof("Basic Authorization Passed. username: %s", user)
	return token.Get(host, user, pwd, scope)
}
