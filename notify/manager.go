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

package notify

import (
	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/notify/provider"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
)

const (
	// Env variables name about EmailNotifier.
	// SMTP_SERVER is the env name for SMTP_SERVER.
	SMTP_SERVER = "SMTP_SERVER"
	// SMTP_PORT is the env name for SMTP_PORT.
	SMTP_PORT = "SMTP_PORT"
	// SMTP_USERNAME is the env name for SMTP_USERNAME.
	SMTP_USERNAME = "SMTP_USERNAME"
	// SMTP_PASSWORD is the env name for SMTP_PASSWORD.
	SMTP_PASSWORD = "SMTP_PASSWORD"
)

// Manager is an asynchronous event manager which records incoming events,
// dispatches events with bookkeeping.
type Manager struct {
	emailNotifier *provider.EmailNotifier
}

var (
	// Env variables about EmailNotifier.
	smtpServer       = osutil.GetStringEnv(SMTP_SERVER, "smtp-default-server")
	smtpPort         = osutil.GetIntEnv(SMTP_PORT, 465)
	smtpUserame      = osutil.GetStringEnv("SMTP_USERNAME", "default-username")
	smtpPassword     = osutil.GetStringEnv(SMTP_PASSWORD, "default-password")
	smtpServerConfig = api.SMTPServerConfig{
		SMTPServer:   smtpServer,
		SMTPPort:     smtpPort,
		SMTPUsername: smtpUserame,
		SMTPPassword: smtpPassword,
	}
)

// newManager return a new NotifyManager, which has a EmailNotifier instance.
func newManager(SMTPServer string, SMTPPort int, SMTPUsername, SMTPPassword string) (*Manager, error) {
	emailNotifier, err := provider.NewEmailNotifier(SMTPServer, SMTPPort, SMTPUsername, SMTPPassword)
	if err != nil {
		return nil, err
	}
	return &Manager{
		emailNotifier: emailNotifier,
	}, nil
}

// Notify sends email to user.
func Notify(service *api.Service, version *api.Version, versionLog string) {
	nManager, err := newManager(smtpServerConfig.SMTPServer, smtpServerConfig.SMTPPort,
		smtpServerConfig.SMTPUsername, smtpServerConfig.SMTPPassword)
	if err != nil {
		log.Warnf("NotifyManager init error: %v", err)
		return
	}
	if err := nManager.emailNotifier.Notify(service, version, versionLog); err != nil {
		log.Warnf("NotifyManager notify error: %v", err)
		return
	}
}

// shouldSendNotifyEvent decides whether cyclone should send notify event.
func shouldSendNotifyEvent(service *api.Service, version *api.Version) bool {
	setting := service.Profile.Setting
	if setting == api.SendWhenFinished {
		return true
	} else if setting == api.SendWhenFailed &&
		(version.Status == api.VersionFailed ||
			version.YamlDeployStatus == api.DeployFailed ||
			deployPlansFailed(version)) {
		return true
	}
	// TODO: Add other settings.
	return false
}

// deployPlansFailed return true if version's deploy plans are all successful.
func deployPlansFailed(version *api.Version) bool {
	for _, plan := range version.DeployPlansStatuses {
		if plan.Status == api.DeployFailed {
			return true
		}
	}
	return false
}
