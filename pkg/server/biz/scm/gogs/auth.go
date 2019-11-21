package gogs

import (
	"fmt"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
)

// LoginInfo post login information to Gogs server to auth cookie
type LoginInfo struct {
	CSRF     string `json:"_csrf"`
	Username string `json:"user_name"`
	Password string `json:"password"`
}

// genCookie generate a auth cookie
func genCookie(scmCfg *v1alpha1.SCMSource) (cookie string, err error) {
	var request = gorequest.New().Timeout(time.Second * 5)
	var response gorequest.Response
	var errs []error

	var url = strings.TrimSuffix(scmCfg.Server, "/")
	if response, _, errs = request.Get(url).End(); len(errs) != 0 {
		err = errs[len(errs)-1]
		return
	}
	if response == nil {
		err = fmt.Errorf("request Gogs server with fatal error: %s", url)
		return
	}
	if response.StatusCode != 200 {
		err = fmt.Errorf("request Gogs server got code: %d, url: %s", response.StatusCode, url)
		return
	}
	for k, v := range response.Header {
		if k == "Set-Cookie" {
			cookie = strings.Join(v, "; ")
			break
		}
	}

	var csrf string
	for _, v := range strings.Split(cookie, "; ") {
		var list = strings.Split(v, "=")
		if len(list) == 2 && list[0] == "_csrf" {
			csrf = list[1]
			break
		}
	}

	var loginInfo = LoginInfo{
		CSRF:     csrf,
		Username: scmCfg.User,
		Password: scmCfg.Password,
	}
	if response, _, errs = request.Post(fmt.Sprintf("%s/user/login", url)).
		SendStruct(loginInfo).
		Type("form").Set("Cookie", cookie).End(); len(errs) != 0 {
		err = errs[len(errs)-1]
		return
	}
	if response == nil {
		err = fmt.Errorf("request Gogs server with fatal error: %s", url)
		return
	}
	if response.StatusCode != 200 {
		err = fmt.Errorf("request Gogs server got code: %d, url: %s", response.StatusCode, url)
		return
	}
	return
}
