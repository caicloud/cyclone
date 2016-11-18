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
	"path"
	"strings"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/executil"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	steplog "github.com/caicloud/cyclone/worker/log"
	"github.com/google/go-github/github"
	gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

// Git is the type for git provider.
type Git struct{}

// NewGit returns a new Git worker.
func NewGit() *Git {
	return &Git{}
}

// CloneRepo implements VCS interface.
func (g *Git) CloneRepo(url, destPath string, event *api.Event) error {
	log.InfoWithFields("About to clone git repository.", log.Fields{"url": url, "destPath": destPath})

	base := path.Base(destPath)
	dir := path.Dir(destPath)
	args := []string{"clone", url, base}

	output, err := executil.RunInDir(dir, "git", args...)
	if event.Version.VersionID != "" {
		fmt.Fprintf(steplog.Output, "%s", string(output))
	}

	if err != nil {
		log.ErrorWithFields("Error when clone", log.Fields{"error": err})
	} else {
		log.InfoWithFields("Successfully cloned git repository.", log.Fields{"url": url, "destPath": destPath})
	}
	return err
}

// NewTagFromLatest implements VCS interface.
func (g *Git) NewTagFromLatest(repoPath string, event *api.Event) error {
	service := event.Service
	version := event.Version
	if service.Repository.SubVcs == api.GITHUB {
		tagName := version.Name + api.AutoCreateTagFlag
		objecttype := "commit"
		curtime := time.Now()
		email := "circle@caicloud.io"
		name := "circle"

		tag := &github.Tag{
			Tag:     &tagName,
			Message: &(version.Description),
			Object: &github.GitObject{
				Type: &objecttype,
				SHA:  &event.Version.Commit,
			},
			Tagger: &github.CommitAuthor{
				Date:  &curtime,
				Name:  &name,
				Email: &email,
			},
		}

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: event.Data["Token"].(string)},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		client := github.NewClient(tc)

		owner, repo := parseURL(service.Repository.URL)
		_, _, err := client.Git.CreateTag(owner, repo, tag)
		if err != nil {
			return err
		}

		ref := "refs/tags/" + tagName
		reference := &github.Reference{
			Ref: &ref,
			Object: &github.GitObject{
				Type: &objecttype,
				SHA:  &event.Version.Commit,
			},
		}
		refs, _, err := client.Git.CreateRef(owner, repo, reference)
		log.Info(refs)
		return err
	} else if service.Repository.SubVcs == api.GITLAB {
		owner, name := parseURL(service.Repository.URL)
		tagname := version.Name + api.AutoCreateTagFlag
		ref := "master"
		tag := &gitlab.CreateTagOptions{
			TagName: &tagname,
			Ref:     &ref,
			Message: &(version.Description),
		}

		gitlabServer := osutil.GetStringEnv("SERVER_GITLAB", "https://gitlab.com")
		client := gitlab.NewOAuthClient(nil, event.Data["Token"].(string))
		client.SetBaseURL(gitlabServer + "/api/v3/")

		_, _, err := client.Tags.CreateTag(owner+"/"+name, tag)
		return err
	}
	return nil
}

// CheckoutTag implements VCS interface.
func (g *Git) CheckoutTag(repoPath string, tag string) error {
	args := []string{"checkout", tag}
	output, err := executil.RunInDir(repoPath, "git", args...)

	// TODO: Need a force checkout if tree is dirty (in what cases could
	// local tree be dirty?
	log.Debugf("Command output: %+v", string(output))
	if err == nil {
		log.InfoWithFields("Successfully checked out to git tag.", log.Fields{"repoPath": repoPath, "tag": tag})
	}
	return err
}

// GetTagCommit implements VCS interface.
func (g *Git) GetTagCommit(repoPath string, tag string) (string, error) {
	args := []string{"rev-list", "-n", "1", tag}
	output, err := executil.RunInDir(repoPath, "git", args...)

	return strings.Trim(string(output), "\n"), err
}

func checkoutMaster(repoPath string) {
	args := []string{"checkout", "master"}
	_, err := executil.RunInDir(repoPath, "git", args...)
	if err != nil {
		log.Warn("checkout to master branch failed.")
	}

	args = []string{"pull", "origin", "master"}
	_, err = executil.RunInDir(repoPath, "git", args...)
	if err != nil {
		log.Warn("update master branch to the lastest failed.")
	}
}

// CheckOutByCommitID checks out code in repo by special commit id.
func (g *Git) CheckOutByCommitID(commitID string, repoPath string, event *api.Event) error {
	log.Infof("checkout commit: %s", commitID)

	args := []string{"-C", repoPath, "reset", "--hard", commitID}

	output, err := executil.RunInDir(repoPath, "git", args...)
	fmt.Fprintf(steplog.Output, "%s", string(output))

	if err != nil {
		log.ErrorWithFields("Error when checkout", log.Fields{"error": err})
	} else {
		log.Info("Successfully checkout commit.")
	}
	return err
}

// parseURL is a helper func to parse the url,such as https://github.com/caicloud/test.git
// to return owner(caicloud) and name(test)
func parseURL(url string) (string, string) {
	strs := strings.SplitN(url, "/", -1)
	name := strings.SplitN(strs[4], ".", -1)
	return strs[3], name[0]
}
