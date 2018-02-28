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
	"path"
	"strings"

	log "github.com/zoumo/logdog"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/executil"
	"github.com/caicloud/cyclone/pkg/worker/scm"
)

// Git is the type for git provider.
type Git struct{}

func init() {
	if err := scm.RegisterProvider(api.GitHub, new(Git)); err != nil {
		log.Error(err)
	}

	if err := scm.RegisterProvider(api.GitLab, new(Git)); err != nil {
		log.Error(err)
	}
}

func (g *Git) Clone(url, destPath string) (string, error) {
	log.Info("About to clone git repository.", log.Fields{"url": url, "destPath": destPath})

	base := path.Base(destPath)
	dir := path.Dir(destPath)
	args := []string{"clone", url, base}

	output, err := executil.RunInDir(dir, "git", args...)

	if err != nil {
		log.Error("Error when clone", log.Fields{"error": err})
	} else {
		log.Info("Successfully cloned git repository.", log.Fields{"url": url, "destPath": destPath})
	}
	return string(output), err
}

// GetTagCommit implements VCS interface.
func (g *Git) GetTagCommit(repoPath string, tag string) (string, error) {
	args := []string{"rev-list", "-n", "1", tag}
	output, err := executil.RunInDir(repoPath, "git", args...)

	return strings.Trim(string(output), "\n"), err
}
