package sonarqube

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// Sonar is the type for sonarqube integration.
type Sonar struct {
	server string
	token  string
}

// NewSonar creates a sonar struct
func NewSonar(url, token string) (*Sonar, error) {
	valid, err := validate(url, token)
	if err != nil || !valid {
		return nil, cerr.ErrorAuthenticationFailed.Error()
	}

	return &Sonar{
		server: url,
		token:  token,
	}, nil
}

// validate validate the token.
func validate(url, token string) (bool, error) {
	url = strings.TrimSuffix(url, "/")
	path := fmt.Sprintf("%s/api/authentication/validate", url)

	log.Infof("test path:%s", path)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return false, err
	}

	// -u your-token: , colon(:) is needed.
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(token+":"))))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Fail to validate sonarqube token as %s", err.Error())
		return false, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Fail to validate sonarqube token as %s", err.Error())
		return false, err
	}

	valid := &validResp{}
	if resp.StatusCode/100 == 2 {
		err := json.Unmarshal(body, valid)
		if err != nil {
			return false, err
		}

		return valid.Valid, nil
	}

	err = fmt.Errorf("Fail to validate sonarqube token as %s, resp code: %v ", body, resp.StatusCode)
	return false, err
}

type validResp struct {
	Valid bool `json:"valid"`
}
