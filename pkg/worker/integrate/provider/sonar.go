package provider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/api"
	executil "github.com/caicloud/cyclone/pkg/util/exec"
	"github.com/caicloud/cyclone/pkg/worker/common"
	"github.com/caicloud/cyclone/pkg/worker/integrate"
	"strings"
)

// sonar is the type for sonarqube integration.
type Sonar struct {
}

func init() {
	if err := integrate.Register(api.IntegrationTypeSonar, new(Sonar)); err != nil {
		log.Error(err)
	}
}

func (s *Sonar) CodeScan(url, token string, config *integrate.CodeScanConfig) (string, error) {
	log.Infof("About to scan code, url:%s", url)
	type cmd struct {
		dir  string
		args []string
	}

	args := []string{
		fmt.Sprintf("-Dsonar.projectName=%s", config.ProjectName),
		fmt.Sprintf("-Dsonar.projectKey=%s", config.ProjectKey),
		fmt.Sprintf("-Dsonar.sources=%s", config.SourcePath),
		fmt.Sprintf("-Dsonar.host.url=%s", url),
		fmt.Sprintf("-Dsonar.login=%s", token),
	}

	for _, arg := range config.ExtensionAgrs {
		args = append(args, arg)
	}

	// sonar-scanner -Dsonar.projectName=cyclone-new -Dsonar.projectKey=cyclone-new -Dsonar.sources=.
	// -Dsonar.host.url=http://192.168.21.100:9000 -Dsonar.login=f399878566d5d6a3de1759222a4b5eb15cac51de
	// -Dsonar.go.coverage.reportPaths=coverage.out
	//
	// -Dsonar.exclusions=vendor/*
	command := cmd{common.CloneDir, args}

	log.Infof("test command: %s", command)
	output, err := executil.RunInDir(command.dir, "sonar-scanner", command.args...)
	if err != nil {
		log.Errorf("Error when clone, command: %s, error: %v, outputs: %s", command, err, string(output))
		return "", err
	}

	log.Info("Successfully scan code by sonar qube, output:%s", string(output))

	return string(output), nil
}

func (s *Sonar) SetCodeScanStatus(url, token string, pid string, status *api.CodeScanStageStatus) error {
	url = strings.TrimSuffix(url, "/")
	if status.SonarQube == nil {
		status.SonarQube = &api.ScanStatusSonarQube{
			OverviewLink: fmt.Sprintf("%s/component_measures?id=%s", url, pid),
		}
	}
	path := fmt.Sprintf("%s/api/measures/component?additionalFields=periods&component=%s"+
		"&metricKeys=reliability_rating,sqale_rating,security_rating,coverage,duplicated_lines_density",
		url, pid)

	log.Infof("test path:%s", path)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	// -u your-token: , colon(colon) is needed.
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(token+":"))))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Fail to get project contents as %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Fail to get sonarqube measures as %s", err.Error())
		return err
	}

	ms := &MeasuresResp{}
	if resp.StatusCode/100 == 2 {
		err := json.Unmarshal(body, ms)
		if err != nil {
			return err
		}

		status.SonarQube.Measures = ms.Component.Measures
		return nil
	}

	err = fmt.Errorf("Fail to get sonarqube measures as %s, resp code: %v ", body, resp.StatusCode)
	return err
}

type MeasuresResp struct {
	Component struct {
		ID        string              `json:"id,omitempty"`
		Key       string              `json:"key,omitempty"`
		Name      string              `json:"name,omitempty"`
		Qualifier string              `json:"qualifier,omitempty"`
		Measures  []*api.SonarMeasure `json:"measures,omitempty"`
	} `json:"component,omitempty"`
}
