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

package stage

import (
	"fmt"
	"strings"
	"time"

	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
	executil "github.com/caicloud/cyclone/pkg/util/exec"
	"github.com/caicloud/cyclone/pkg/worker/cycloneserver"
)

const (
	startLogTmpl           = "Stage: %s status: start\n"
	finishLogTmpl          = "Stage: %s status: finish\n"
	finishWithErrorLogTmpl = "Stage: %s status: fail with error: %v\n"
)

var stageDesps = map[api.PipelineStageName]string{
	api.CodeCheckoutStageName:    "Code checkout",
	api.PackageStageName:         "Package",
	api.CodeScanStageName:        "Code scan",
	api.ImageBuildStageName:      "Build image",
	api.IntegrationTestStageName: "Integration test",
	api.ImageReleaseStageName:    "Push image",
}

func generateStageStartLog(stage api.PipelineStageName) string {
	desp, ok := stageDesps[stage]
	if !ok {
		log.Errorf("not support stage: %s", stage)
		return ""
	}

	return fmt.Sprintf(startLogTmpl, desp)
}

func generateStageFinishLog(stage api.PipelineStageName, err error) string {
	desp, ok := stageDesps[stage]
	if !ok {
		log.Errorf("not support stage: %s", stage)
		return ""
	}

	if err != nil {
		return fmt.Sprintf(finishWithErrorLogTmpl, desp, err)
	}

	return fmt.Sprintf(finishLogTmpl, desp)
}

func updateRecordStageStatus(stages *api.BuildStages, pipelineRecord *api.PipelineRecord, stage api.PipelineStageName, status api.Status, failErr error) error {
	var gss *api.GeneralStageStatus
	allowFailure := false
	stageStatus := pipelineRecord.StageStatus

	switch stage {
	case api.CodeCheckoutStageName:
		if stageStatus.CodeCheckout == nil {
			stageStatus.CodeCheckout = &api.CodeCheckoutStageStatus{
				GeneralStageStatus: api.GeneralStageStatus{},
			}
		}
		gss = &stageStatus.CodeCheckout.GeneralStageStatus
	case api.PackageStageName:
		if stageStatus.Package == nil {
			stageStatus.Package = &api.GeneralStageStatus{}
		}
		gss = stageStatus.Package
		allowFailure = stages.Package.AllowFailure
	case api.CodeScanStageName:
		if stageStatus.CodeScan == nil {
			stageStatus.CodeScan = &api.CodeScanStageStatus{}
		}
		gss = &stageStatus.CodeScan.GeneralStageStatus
		allowFailure = stages.CodeScan.AllowFailure
	case api.ImageBuildStageName:
		if stageStatus.ImageBuild == nil {
			stageStatus.ImageBuild = &api.ImageBuildStageStatus{}
		}
		gss = &stageStatus.ImageBuild.GeneralStageStatus
	case api.IntegrationTestStageName:
		if stageStatus.IntegrationTest == nil {
			stageStatus.IntegrationTest = &api.GeneralStageStatus{}
		}
		gss = stageStatus.IntegrationTest
		allowFailure = stages.IntegrationTest.AllowFailure
	case api.ImageReleaseStageName:
		if stageStatus.ImageRelease == nil {
			stageStatus.ImageRelease = &api.ImageReleaseStageStatus{
				GeneralStageStatus: api.GeneralStageStatus{},
			}
		}
		gss = &stageStatus.ImageRelease.GeneralStageStatus
		allowFailure = stages.ImageRelease.AllowFailure
	default:
		err := fmt.Errorf("stage %s is not supported", stage)
		log.Error(err)
		return err
	}

	switch status {
	case api.Pending:
		gss.Status = api.Pending
	case api.Running:
		gss.Status = api.Running
		gss.StartTime = time.Now()
	case api.Success:
		gss.Status = api.Success
		gss.EndTime = time.Now()
	case api.Failed:
		gss.Status = api.Failed
		gss.EndTime = time.Now()

		if !allowFailure {
			pipelineRecord.Status = api.Failed
		}

		if failErr != nil {
			desp, ok := stageDesps[stage]
			if !ok {
				err := fmt.Errorf("stage status %s is not supported", status)
				log.Error(err)
				return err
			}
			pipelineRecord.ErrorMessage = fmt.Sprintf("%s fails : %v", desp, failErr)
		}

		// Wait for a while to ensure that stage logs are reported to server.
		// The worker will be terminated as soon as the failed status is reported to server.
		time.Sleep(waitTime)
	default:
		err := fmt.Errorf("stage status %s is not supported", status)
		log.Error(err)
		return err
	}

	return nil
}

func updateEvent(c cycloneserver.CycloneServerClient, event *api.Event, stage api.PipelineStageName, status api.Status, failErr error) error {
	if err := updateRecordStageStatus(event.Pipeline.Build.Stages, event.PipelineRecord, stage, status, failErr); err != nil {
		return err
	}

	return c.SendEvent(event)
}

// findGoCoverprofile find golang coverprofile from commands.
func findGoCoverprofile(commands []string) string {
	var file string
	for _, command := range commands {
		// 'go test ... -coverprofile=xxx ...' or 'go test ... -coverprofile xxx ...'
		if strings.Contains(command, "go test") && strings.Contains(command, "-coverprofile") {
			i := strings.Index(command, "-coverprofile")
			tmp := command[i+len("-coverprofile"):]
			commandFields := strings.Fields(tmp)
			file = strings.TrimPrefix(commandFields[0], "=")
		}
	}
	return file
}

// CopyFile cp file from src to dest.
func CopyFile(dir, src, dest string) (string, error) {
	args := []string{src, dest}
	output, err := executil.RunInDir(dir, "cp", args...)
	return string(output), err
}
