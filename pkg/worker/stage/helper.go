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

	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
)

const (
	startLogTmpl           = "Stage: %s status: start\n"
	finishLogTmpl          = "Stage: %s status: finish\n"
	finishWithErrorLogTmpl = "Stage: %s status: fail with error: %v\n"
)

var stageDesps = map[api.PipelineStageName]string{
	api.CodeCheckoutStageName:    "Code checkout",
	api.PackageStageName:         "Package",
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
