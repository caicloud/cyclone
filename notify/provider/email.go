/*
Copyright 2016 caicloud authors. All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/buildkite/terminal"
	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/osutil"
	"gopkg.in/gomail.v2"
)

const (
	// Subject is a string:
	// Successful or Errored: <service name>.
	Subject = "%s: %s"
	// From defines the address of the sender.
	From = "Caicloud CI <%s>"
	// ContentType defines the contentType of the email.
	ContentType = "text/html"
	// WebServiceURL defines the url of cyclone.
	// TODO: Move the definitions of cyclone Port in main to utils package and remove this const var.
	WebServiceURL = "http://localhost:7099"
	// SUCCESSTEMPLATE defines the env var, which points to the template.
	SUCCESSTEMPLATE = "SUCCESSTEMPLATE"
	// ERRORTEMPLATE defines the env var, which points to the template.
	ERRORTEMPLATE = "ERRORTEMPLATE"
	// UsernameVarName defines the username in the template.
	UsernameVarName = "$$USERNAME"
	// ServiceNameVarName defines the service name in the template.
	ServiceNameVarName = "$$SERVICENAME"
	// RepoURLVarName defines the URL of the repo in the template.
	RepoURLVarName = "$$REPOURL"
	// CommitIDVarName defines the commitID in the template.
	CommitIDVarName = "$$COMMITID"
	// LogVarName defines the log name in the template.
	LogVarName = "$$LOG"
	// DeployStatusVarName defines the deploy status in the template.
	DeployStatusVarName = "$$DEPLOYSTATUS"
	// DeployPlanStatusVarName defines the deploy plan status in the template.
	DeployPlanStatusVarName = "$$DEPLOYPLANSTATUS"
)

var (
	// SuccessFilename is the file name for SUCCESSTEMPLATE.
	SuccessFilename string
	// ErrorFilename is the file name for ERRORTEMPLATE.
	ErrorFilename string
	// SuccessfulBody defines the body of the email when the version is successful.
	SuccessfulBody string
	// ErroredBody defines the body of the email when the version is failed.
	ErroredBody string
)

// EmailNotifier is the email Notifier, which uses the Paging Client to send emails to users.
type EmailNotifier struct {
	SMTPServer   string
	SMTPPort     int
	SMTPUsername string
	SMTPPassWord string
}

// NewEmailNotifier returns a new EmailNotifier.
func NewEmailNotifier(SMTPServer string, SMTPPort int, SMTPUsername, SMTPPassword string) (*EmailNotifier, error) {
	// Read the HTML template from the file.
	err := readContextFromConfigFile()
	if err != nil {
		return nil, err
	}
	return &EmailNotifier{
		SMTPServer:   SMTPServer,
		SMTPPort:     SMTPPort,
		SMTPUsername: SMTPUsername,
		SMTPPassWord: SMTPPassword,
	}, nil
}

// readContextFromConfigFile reads the HTML template from files.
func readContextFromConfigFile() error {
	SuccessFilename = osutil.GetStringEnv(SUCCESSTEMPLATE, "/template/success.html")
	ErrorFilename = osutil.GetStringEnv(ERRORTEMPLATE, "/template/error.html")
	context, err := ioutil.ReadFile(SuccessFilename)
	if err != nil {
		return err
	}
	SuccessfulBody = string(context)

	context, err = ioutil.ReadFile(ErrorFilename)
	if err != nil {
		return err
	}
	ErroredBody = string(context)
	return nil
}

// Notify sends the email to the email address defined in the api.Service.
func (e EmailNotifier) Notify(service *api.Service, version *api.Version, versionLog string) error {
	var body string

	// Initialize the content of the email.
	subject := fmt.Sprintf(Subject, version.Status, service.Name)
	from := fmt.Sprintf(From, e.SMTPUsername)
	if version.Status == api.VersionHealthy {
		body = fillUpTemplate(service, version, versionLog, SuccessfulBody)
	} else {
		body = fillUpTemplate(service, version, versionLog, ErroredBody)
	}

	// Create the new email.
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	for _, p := range service.Profile.Profiles {
		m.SetHeader("To", p.Mail)
	}
	m.SetHeader("Subject", subject)
	m.SetBody(ContentType, body)

	d := gomail.NewDialer(e.SMTPServer, e.SMTPPort, e.SMTPUsername, e.SMTPPassWord)
	// Send the email from the SMTPServer:SMTPPort.
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// fillUpTemplate push the var to template, and returns a HTML string.
func fillUpTemplate(service *api.Service, version *api.Version, logInANSI, bodyTemplate string) string {
	var body string
	// Render the ANSI log to HTML format.
	logInHTML := string(terminal.Render([]byte(logInANSI)))
	// Replace the template with the variables.
	body = strings.Replace(bodyTemplate, UsernameVarName, service.Username, -1)
	body = strings.Replace(body, RepoURLVarName, service.Repository.URL, -1)
	body = strings.Replace(body, CommitIDVarName, version.Commit, -1)
	body = strings.Replace(body, DeployStatusVarName, string(version.YamlDeployStatus), -1)
	planstatus := ""
	for _, plan := range version.DeployPlansStatuses {
		planstatus += fmt.Sprintf("%s:%s ", plan.PlanName, string(plan.Status))
	}
	if planstatus != "" {
		body = strings.Replace(body, DeployPlanStatusVarName, planstatus, -1)
	} else {
		body = strings.Replace(body, DeployPlanStatusVarName, string(api.DeployNoRun), -1)
	}
	body = strings.Replace(body, ServiceNameVarName, service.Name, -1)
	body = strings.Replace(body, LogVarName, logInHTML, -1)
	return body
}
