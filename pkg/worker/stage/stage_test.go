package stage_test

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/docker"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/worker/cycloneserver"
	_ "github.com/caicloud/cyclone/pkg/worker/scm/provider"
	. "github.com/caicloud/cyclone/pkg/worker/stage"
)

const (
	logDir  = "/tmp/logs"
	codeDir = "/tmp/code"
)

var stageManager StageManager

func init() {
	// Log to standard error instead of files.
	flag.Set("logtostderr", "true")

	// Init the common clients.
	dm, err := docker.NewDockerManager("unix:////Users/robin/Library/Containers/com.docker.docker/Data/s60", "", "", "")
	if err != nil {
		panic(err)
	}

	client := cycloneserver.NewFakeClient("http://fack-server.cyclone.com")
	event := &api.Event{
		ID:       "1",
		Project:  &api.Project{},
		Pipeline: &api.Pipeline{},
		PipelineRecord: &api.PipelineRecord{
			PerformParams: &api.PipelinePerformParams{
				Ref: "master",
			},
			StageStatus: &api.StageStatus{},
		},
	}
	stageManager = NewStageManager(dm, client, event.PipelineRecord.PerformParams)
	stageManager.SetEvent(event)

	fmt.Println("Initialization of common clients has finished")
}

func TestExecCodeCheckout(t *testing.T) {
	testCases := map[string]struct {
		inputs []*api.CodeSource
		pass   bool
	}{
		"correct public github": {
			[]*api.CodeSource{
				&api.CodeSource{
					Type: api.GitHub,
					GitHub: &api.GitSource{
						Url: "https://github.com/caicloud/toy-dockerfile.git",
					},
				},
			},
			true,
		},
		"correct private github": {
			[]*api.CodeSource{
				&api.CodeSource{
					Type: api.GitHub,
					GitHub: &api.GitSource{
						Url: "https://github.com/caicloud/dockerfile.git",
					},
				},
			},
			false,
		},
		"wrong github": {
			[]*api.CodeSource{
				&api.CodeSource{
					Type: api.GitHub,
					GitHub: &api.GitSource{
						Url: "https://github.com/caicloud/abc.git",
					},
				},
			},
			false,
		},
	}

	stage := &api.CodeCheckoutStage{}
	for d, tc := range testCases {
		// Cleanup the temp folder.
		os.RemoveAll(codeDir)
		stage.CodeSources = tc.inputs
		err := stageManager.ExecCodeCheckout("", stage)
		if tc.pass && err != nil || !tc.pass && err == nil {
			t.Errorf("%s failed as error: %v", d, err)
		}
	}
}

func TestExecImageBuild(t *testing.T) {
	// Prepare the code folder.
	dockerfileContent := "FROM alpine \nADD README.md /README.md"
	readmeContent := "Hello Cyclone"
	os.RemoveAll(codeDir)
	os.MkdirAll(codeDir, os.ModePerm)
	osutil.ReplaceFile(codeDir+"/Dockerfile", strings.NewReader(dockerfileContent))
	osutil.ReplaceFile(codeDir+"/README.md", strings.NewReader(readmeContent))

	testCases := map[string]struct {
		inputs []*api.ImageBuildInfo
		pass   bool
	}{
		"default Dockerfile": {
			[]*api.ImageBuildInfo{
				&api.ImageBuildInfo{
					ImageName: "cargo.caicloud.io/caicloud/test:v1",
				},
			},
			true,
		},
		"correct Dockerfile path": {
			[]*api.ImageBuildInfo{
				&api.ImageBuildInfo{
					ImageName:      "cargo.caicloud.io/caicloud/test:v1",
					DockerfilePath: "Dockerfile",
				},
			},
			true,
		},
		"wrong Dockerfile path": {
			[]*api.ImageBuildInfo{
				&api.ImageBuildInfo{
					ImageName:      "cargo.caicloud.io/caicloud/test:v1",
					DockerfilePath: "Dockerfile-abc",
				},
			},
			false,
		},
		"correct Dockerfile content": {
			[]*api.ImageBuildInfo{
				&api.ImageBuildInfo{
					ImageName:  "cargo.caicloud.io/caicloud/test:v1",
					Dockerfile: "FROM alpine \nADD README.md /README.md",
				},
			},
			true,
		},
		// "wrong Dockerfile content": {
		// 	[]*api.ImageBuildInfo{
		// 		&api.ImageBuildInfo{
		// 			ImageName:  "cargo.caicloud.io/caicloud/test:v1",
		// 			Dockerfile: "FROM alpine \nADD readme.md /readme.md",
		// 		},
		// 	},
		// 	false,
		// },
	}

	stage := &api.ImageBuildStage{}
	for d, tc := range testCases {
		stage.BuildInfos = tc.inputs
		_, err := stageManager.ExecImageBuild(stage)
		if tc.pass && err != nil || !tc.pass && err == nil {
			t.Errorf("%s failed as error: %v", d, err)
		}
	}
}
