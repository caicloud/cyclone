package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/util/oauth"
	"github.com/caicloud/cyclone/pkg/util/os"

	"github.com/caicloud/nirvana/service"
)

// GetAuthCodeURL get the redirect url from config
// then return url to frontend
func GetAuthCodeURL(ctx context.Context, scm string) (string, error) {
	client, err := oauth.NewOauthClient(scm)
	if err != nil {
		return "", err
	}

	return client.GetAuthCodeURL(oauth.FakeState), nil
}

// GetToken use code to change token and redirect to frontend
func GetToken(ctx context.Context, scm string, code string, state string) error {
	client, err := oauth.NewOauthClient(scm)
	if err != nil {
		return err
	}

	// get oauth token(AccessToken)
	token, err := client.GetToken(state, code)
	if err != nil {
		return err
	}
	username, server, err := client.GetUserInfo(token)
	if err != nil {
		return err
	}
	consoleWebServer := os.GetStringEnv(options.ConsoleWebEndpoint, "")
	redirectURL := fmt.Sprintf(
		"%s/devops/workspace/add?type=gitlab&code=%s&state=%s&token=%s&username=%s&server=%s",
		consoleWebServer, code, state, token, username, server,
	)

	httpCtx := service.HTTPContextFrom(ctx)
	w := httpCtx.ResponseWriter()
	r := httpCtx.Request()
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	return nil
}
