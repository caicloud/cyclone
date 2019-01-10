package sonar

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/integrate"
	"github.com/caicloud/cyclone/pkg/junit"
	executil "github.com/caicloud/cyclone/pkg/util/exec"
)

const (
	warExt         = ".war"
	jarExt         = ".jar"
	DirVendor      = "vendor"
	DirNodeModules = "node_modules"

	// ceTaskIdPrefix is the prefix of the line in  report-task.txt file.
	ceTaskIdPrefix = "ceTaskId="
)

// When sonnar-scanner executed successfully, there will be an report file in
// '/tmp/code/.scannerwork/report-task.txt'
var reportFilePath = filepath.Join(common.CloneDir, ".scannerwork/report-task.txt")

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

	// set the default value of sources to '.'
	sourcePath := "."
	if config.SourcePath != "" {
		sourcePath = config.SourcePath
	}

	args := []string{
		fmt.Sprintf("-Dsonar.projectName=%s", config.ProjectName),
		fmt.Sprintf("-Dsonar.projectKey=%s", config.ProjectKey),
		fmt.Sprintf("-Dsonar.sources=%s", sourcePath),
		fmt.Sprintf("-Dsonar.host.url=%s", url),
		fmt.Sprintf("-Dsonar.login=%s", token),
		fmt.Sprintf("-Dsonar.exclusions=%s/***,%s/***", DirVendor, DirNodeModules),
	}

	// Find java binary files.
	javaBinaries := findJavaBinaryFiles(common.CloneDir)
	if len(javaBinaries) > 0 {
		args = append(args, fmt.Sprintf("-Dsonar.java.binaries=%s", strings.Join(javaBinaries, ",")))
	}

	// Find report files
	var reportFiles []string

	// Go HTML Java JavaScript PHP Python
	switch config.Language {
	case api.GolangType:
		// Find golang tests.reportPaths is a little tricky, we will try our best to analysis commands in unit-test stage
		// to find out where the test report is, and add copy the file to go_test_report.cyclone, so we can use it here.
		args = append(args, fmt.Sprintf("-Dsonar.go.coverage.reportPaths=%s", common.GoTestReport))
	case api.JavaRepoType:
		report := junit.NewReport(common.CloneDir)
		reportFiles = report.FindReportFiles()
		args = append(args, fmt.Sprintf("-Dsonar.junit.reportPaths=%s", strings.Join(reportFiles, ",")))
	default:
		log.Info("coverage.reportPaths language:%s have not supported.", config.Language)
	}

	// Add extension args.
	for _, arg := range config.ExtensionAgrs {
		args = append(args, arg)
	}

	// sonar-scanner -Dsonar.projectName=cyclone-new -Dsonar.projectKey=cyclone-new -Dsonar.sources=.
	// -Dsonar.host.url=http://192.168.21.100:9000 -Dsonar.login=f399878566d5d6a3de1759222a4b5eb15cac51de
	//
	// -Dsonar.exclusions=vendor/***,node_modules/*** -Dsonar.go.coverage.reportPaths=coverage.out -Dsonar.java.binaries=xx.jar
	command := cmd{common.CloneDir, args}

	log.Infof("test command: %s", command)
	output, err := executil.RunInDir(command.dir, "sonar-scanner", command.args...)
	if err != nil {
		log.Errorf("Error when code scan, command: %s, error: %v, outputs: %s", command, err, string(output))
		return "", err
	}

	log.Info("Successfully scan code by sonar qube, output:%s", string(output))

	return string(output), nil
}

// extractCeTaskId extracts compute engine task id.
//
// When sonnar-scanner executed successfully, there will be an file in
// '.scannerwork/report-task.txt' which containers following info:
// '......
//  ceTaskId=AWd9EPiuomqmtD7hc2iT
//  ceTaskUrl=http://192.168.21.100:9000/api/ce/task?id=AWd9EPiuomqmtD7hc2iT
// '
// And we need the ceTaskId 'AWd9EPiuomqmtD7hc2iT' to wait the analysis to be complete.
func extractCeTaskId(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Error("open file %s failed", path)
		return "", err
	}
	defer file.Close()

	var id string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), ceTaskIdPrefix) {
			id = strings.TrimPrefix(scanner.Text(), ceTaskIdPrefix)
			return strings.TrimSpace(id), nil
		}
	}

	return "", fmt.Errorf("ceTaskId not found")
}

func (s *Sonar) SetCodeScanStatus(url, token string, pid string, status *api.CodeScanStageStatus) error {
	url = strings.TrimSuffix(url, "/")
	if status.SonarQube == nil {
		status.SonarQube = &api.ScanStatusSonarQube{
			OverviewLink: fmt.Sprintf("%s/dashboard?id=%s", url, pid),
		}
	}

	// extract ceTaskId
	ceTaskId, err := extractCeTaskId(reportFilePath)
	if err != nil {
		log.Error("extract ceTaskId failed, error:%v", err)
		return err
	}

	completed := false
	// wait for sonarqube server to complete this analysis task.
	for i := 0; i < 10; i++ {
		cet, err := getCeTask(url, token, ceTaskId)
		if err != nil {
			log.Warningf("get compute engine task id failed: %v", err)
		} else {
			if cet.Task.Status == "SUCCESS" || cet.Task.Status == "FAILED" {
				log.Infof("ce task %s complete, status: %s", ceTaskId, cet.Task.Status)
				completed = true
				break
			}
		}

		time.Sleep(6 * time.Second)
	}

	if !completed {
		return fmt.Errorf("wait for sonarqube server to complete this analysis task, time out")
	}

	// get project measures
	ms, err := getProjectMeasures(url, token, pid)
	if err != nil {
		return err
	}
	status.SonarQube.Measures = ms.Component.Measures

	return nil
}

func getProjectMeasures(url, token string, pid string) (*MeasuresResp, error) {
	path := fmt.Sprintf("%s/api/measures/component?additionalFields=periods&component=%s"+
		"&metricKeys=reliability_rating,sqale_rating,security_rating,coverage,duplicated_lines_density,quality_gate_details",
		url, pid)

	log.Infof("test path:%s", path)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// -u your-token: , colon(:) is needed.
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(token+":"))))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Fail to get sonarqube measures as %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Fail to get sonarqube measures as %s", err.Error())
		return nil, err
	}

	ms := &MeasuresResp{}
	if resp.StatusCode/100 == 2 {
		err := json.Unmarshal(body, ms)
		if err != nil {
			return nil, err
		}

		return ms, nil
	}

	err = fmt.Errorf("Fail to get sonarqube measures as %s, resp code: %v ", body, resp.StatusCode)
	return nil, err
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

// getCeTask gets Compute Engine task details.
func getCeTask(url, token string, id string) (*CeTaskResp, error) {
	path := fmt.Sprintf("%s/api/ce/task?id=%s", url, id)

	log.Infof("test path:%s", path)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// -u your-token: , colon(:) is needed.
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(token+":"))))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Fail to get sonarqube measures as %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Fail to get sonarqube measures as %s", err.Error())
		return nil, err
	}

	cet := &CeTaskResp{}
	if resp.StatusCode/100 == 2 {
		err := json.Unmarshal(body, cet)
		if err != nil {
			return nil, err
		}

		return cet, nil
	}

	err = fmt.Errorf("Fail to get sonarqube measures as %s, resp code: %v ", body, resp.StatusCode)
	return nil, err
}

type CeTaskResp struct {
	Task struct {
		ID     string `json:"id,omitempty"`
		Type   string `json:"type,omitempty"`
		Status string `json:"status,omitempty"`
	} `json:"task,omitempty"`
}

func findJavaBinaryFiles(root string) []string {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil
	}

	var files []string
	filepath.Walk(root, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			// Exclude vendor and node_modules dir.
			if strings.HasPrefix(path, DirVendor) || strings.HasPrefix(path, DirNodeModules) {
				return nil
			}

			// Java binary files end with '.jar' or '.war'
			if (filepath.Ext(path) != warExt) && (filepath.Ext(path) != jarExt) {
				return nil
			}

			files = append(files, path)
		}
		return nil
	})

	return files
}

// CreateProject create a project.
func (s *Sonar) CreateProject(url, token string, projectKey, projectName string) error {
	url = strings.TrimSuffix(url, "/")
	path := fmt.Sprintf("%s/api/projects/create?project=%s&name=%s",
		url, projectKey, projectName)

	log.Infof("test path:%s", path)
	req, err := http.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	// -u your-token: , colon(:) is needed.
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(token+":"))))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Fail to create sonarqube project as %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Fail to create sonarqube project as %s", err.Error())
		return err
	}

	if resp.StatusCode/100 == 2 {
		return nil
	}

	err = fmt.Errorf("Fail to create sonarqube project as %s, resp code: %v ", body, resp.StatusCode)
	return err
}

// SetQualityGate sets the project's quality gate.
func (s *Sonar) SetQualityGate(url, token string, projectKey string, gateId int) error {
	url = strings.TrimSuffix(url, "/")
	path := fmt.Sprintf("%s/api/qualitygates/select?projectKey=%s&gateId=%d",
		url, projectKey, gateId)

	log.Infof("test path:%s", path)
	req, err := http.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	// -u your-token: , colon(:) is needed.
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(token+":"))))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Fail to set quality gate for project %s as %s", projectKey, err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Fail to set quality gate for project %s as %s", projectKey, err.Error())
		return err
	}

	if resp.StatusCode/100 == 2 {
		return nil
	}

	err = fmt.Errorf("Fail to set quality gate for project %s as %s, resp code: %v ",
		projectKey, body, resp.StatusCode)
	return err
}

// DeleteProject delete a project.
func (s *Sonar) DeleteProject(url, token string, projectKey string) error {
	url = strings.TrimSuffix(url, "/")
	path := fmt.Sprintf("%s/api/projects/delete?project=%s", url, projectKey)

	log.Infof("test path:%s", path)
	req, err := http.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	// -u your-token: , colon(:) is needed.
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(token+":"))))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Fail to delete sonarqube project as %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Fail to delete sonarqube project as %s", err.Error())
		return err
	}

	if resp.StatusCode/100 == 2 {
		return nil
	}

	err = fmt.Errorf("Fail to delete sonarqube project as %s, resp code: %v ", body, resp.StatusCode)
	return err
}

// Validate validate the token.
func (s *Sonar) Validate(url, token string) (bool, error) {
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

	valid := &ValidResp{}
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

type ValidResp struct {
	Valid bool `json:"valid"`
}
