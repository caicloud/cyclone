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

package rest

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/executil"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"
	"github.com/emicklei/go-restful"
	"github.com/satori/go.uuid"
)

// webhookGithub handles webhook data from github.
//
// POST: /api/v0.1/{service_id}/webhook_github
//
// RESPONSE: (WebhookResponse)
//  {
//    "error_msg": (string) set IFF the request fails.
//  }
func webhookGithub(request *restful.Request, response *restful.Response) {
	// Read payload.
	payload := api.WebhookGithub{}
	err := request.ReadEntity(&payload)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	var webhookResponse api.WebhookResponse
	serviceID := request.PathParameter("service_id")

	// Check vcs of service is github.
	if api.Git != getServiceVcs(serviceID) {
		webhookResponse.ErrorMessage = "vcs is not github"
		response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
		return
	}

	// Distinguish event type according to the payload.
	eventType := request.HeaderParameter("X-GitHub-Event")

	// Ignore repeated events.
	if isNeedIgnore(eventType, payload) {
		log.Info("ignore repeated events.")
		webhookResponse.ErrorMessage = "ignore"
		response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
		return
	}

	// Set create version operation:
	// 	commit: integration
	// 	tag: integration and pblish
	// 	pull request: integration
	var operation api.VersionOperation
	if isReleaseTag(eventType, payload) {
		log.Info("It's a release tag event")
		operation = api.IntegrationOperation + api.PublishOperation + api.DeployOperation
	} else {
		operation = api.IntegrationOperation
	}

	// Create version config.
	version := createVersionGithubConfig(serviceID, eventType, payload, operation)
	if nil == version {
		webhookResponse.ErrorMessage = "unknow"
		response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
		return
	}

	// Create a version.
	errCreateVersion := webhookCreateVersion(serviceID, version)
	if nil != errCreateVersion {
		webhookResponse.ErrorMessage = errCreateVersion.Error()
		response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
		return
	}

	webhookResponse.ErrorMessage = "ok"
	response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
}

// getServiceVcs gets vcs type of the given ServiceID.
func getServiceVcs(serviceID string) api.VersionControlSystem {
	// find service info from DB
	ds := store.NewStore()
	defer ds.Close()

	service, err := ds.FindServiceByID(serviceID)
	if err != nil {
		return ""
	}
	return service.Repository.Vcs
}

// isNeedIgnore returns if the event needs to be ignored according to the payload.
// Now, we receive 2 types events: push, PR in github.
func isNeedIgnore(eventType string, payload api.WebhookGithub) bool {
	if api.GithubWebhookPush == eventType {
		// Create a version in UI, it will create a tag in repo automatically.
		// Need to ignore it.
		if isAutoCreateReleaseTag(payload) {
			return true
		}
	} else if api.GithubWebhookPullRequest == eventType {
		if payload[api.GithubWebhookFlagAction] != api.PRActionOpened &&
			payload[api.GithubWebhookFlagAction] != api.PRActionSynchronize {
			return true
		}
	} else {
		// Do Nothing.
	}

	return false
}

// isReleaseTag returns if the event is a release tag event according to the payload.
func isReleaseTag(eventType string, payload api.WebhookGithub) bool {
	if api.GithubWebhookPush == eventType {
		if nil != payload[api.GithubWebhookFlagRef] &&
			strings.Contains(payload[api.GithubWebhookFlagRef].(string),
				api.GithubWebhookFlagTags) {
			return true
		}
	}

	return false
}

// AutoCreateTagFlag returns if the event is a auto create release tag event
// according to the payload.
func isAutoCreateReleaseTag(payload api.WebhookGithub) bool {
	if nil != payload[api.GithubWebhookFlagRef] &&
		strings.Contains(payload[api.GithubWebhookFlagRef].(string), api.GithubWebhookFlagTags) &&
		strings.HasSuffix(payload[api.GithubWebhookFlagRef].(string), api.AutoCreateTagFlag) {
		return true
	}
	return false
}

// createVersionGithubConfig creates a version config.
func createVersionGithubConfig(serviceID string, eventType string, payload api.WebhookGithub,
	operation api.VersionOperation) *api.Version {
	version := &api.Version{
		ServiceID:        serviceID,
		Operation:        operation,
		YamlDeploy:       api.DeployWithYaml,
		YamlDeployStatus: api.DeployNoRun,
		SecurityCheck:    false,
	}

	switch eventType {
	case api.GithubWebhookPush:
		log.Info("webhook event: push")
		version.Name, version.Description, version.Commit = generateVersionFromPushData(payload)

	case api.GithubWebhookPullRequest:
		log.Info("webhook event: pull_request")
		version.Name, version.Description, version.URL, version.Commit = generateVersionFromPRData(payload)

	default:
		log.Info("receive undefine webhook event")
		return nil
	}

	return version
}

// generateVersionFromPushData generates version config from payload data.
// name:
//   tag: tag_commitId
//   commit: ci_commitId
// description: message commitId
func generateVersionFromPushData(payload api.WebhookGithub) (name string, description string,
	commitID string) {
	commit := payload[api.GithubWebhookFlagCommit].(map[string]interface{})
	ref := payload[api.GithubWebhookFlagRef].(string)

	// generate name
	var namePrefix string
	var tagName string
	if isReleaseTag(api.GithubWebhookPush, payload) {
		namePrefix = "tag_"
		// e.g. refs/tags/v0.12
		tagName = strings.SplitN(ref, "/", -1)[2]
	} else {
		namePrefix = "ci_"
	}

	if nil != commit["id"] {
		if "" != tagName && canBeUsedInImageTag(tagName) {
			name = namePrefix + tagName
		} else {
			name = namePrefix + commit["id"].(string)
		}
	} else {
		name = namePrefix + uuid.NewV4().String()
	}

	// generate description
	if nil != commit["message"] {
		description = commit["message"].(string)
	} else {
		description = ""
	}
	if nil != commit["id"] {
		commitID = commit["id"].(string)
		description = description + "\r\n" + commit["id"].(string)
	}

	log.Infof("webhook push event: name[%s] description[%s] commit[%s]", name, description, commitID)
	return name, description, commitID
}

// generateVersionFromPRData generates version config from payload data.
// name: prNo_commitId
// description: title body commitId pr_url
func generateVersionFromPRData(payload api.WebhookGithub) (name string, description string,
	url string, commitID string) {
	pullRequest := payload[api.GithubWebhookFlagPR].(map[string]interface{})

	// Analyse PR url.
	head := pullRequest["head"].(map[string]interface{})
	repo := head["repo"].(map[string]interface{})
	url = repo["clone_url"].(string)

	// Analyse PR commit.
	commitID = head["sha"].(string)

	// Generate name.
	if nil != pullRequest["number"] {
		name = fmt.Sprintf("pr_%v_%s", pullRequest["number"], commitID)
	} else {
		name = fmt.Sprintf("pr_%s", commitID)
	}

	// Generate description.
	if nil != pullRequest["title"] {
		description = pullRequest["title"].(string)
	} else {
		description = ""
	}
	if nil != pullRequest["body"] && "" != pullRequest["body"] {
		description = description + "\r\n" + pullRequest["body"].(string)
	}
	description = description + "\r\n" + commitID + "\r\n" + url

	log.Infof("webhook pr event: name[%s] description[%s] url[%s] commit[%s]", name, description,
		url, commitID)
	return name, description, url, commitID
}

// generateVersionFromReleaseData generates version config from payload data.
func generateVersionFromReleaseData(payload api.WebhookGithub) (name string, description string) {
	release := payload[api.GithubWebhookFlagRelease].(map[string]interface{})

	// generate name
	if nil != release["tag_name"] && canBeUsedInImageTag(release["tag_name"].(string)) {
		name = release["tag_name"].(string)
	} else {
		name = "tag_" + uuid.NewV4().String()
	}

	// generate description
	if nil != release["body"] {
		description = release["body"].(string)
	} else {
		description = ""
	}
	if nil != release["html_url"] {
		description = description + "\r\n" + release["html_url"].(string)
	}

	log.Infof("webhook release event: name[%s] description[%s]", name, description)
	return name, description
}

// canBeUsedInImageTag gets if the image's tag is valid.
func canBeUsedInImageTag(str string) bool {
	reg := fmt.Sprintf("[0-9a-zA-Z,-,_,.]{%d}", len(str))
	canBeUesed, _ := regexp.MatchString(reg, str)
	if !canBeUesed {
		log.Infof("[%s] can't be used in image tag", str)
	}
	return canBeUesed
}

// webhookSVN handles webhook data from svn.
//
// POST: /api/v0.1/{service_id}/webhooksvn
//
// PAYLOAD (WenhookSVN):
//   {
//     "url": (string) svn url
//     "event": (string) event type of webhook
//     "commiy_id": (string) id of commit
//   }
//
// RESPONSE: (WebhookResponse)
//  {
//    "error_msg": (string) set IFF the request fails.
//  }
func webhookSVN(request *restful.Request, response *restful.Response) {
	// read payload
	payload := api.WebhookSVN{}
	err := request.ReadEntity(&payload)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	var webhookResponse api.WebhookResponse
	serviceID := request.PathParameter("service_id")

	// Make sure that the vcs of service is svn.
	ds := store.NewStore()
	defer ds.Close()
	service, err := ds.FindServiceByID(serviceID)
	if err != nil {
		webhookResponse.ErrorMessage = "unknow serviceID"
		response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
		return
	}

	if api.Svn != service.Repository.Vcs {
		webhookResponse.ErrorMessage = "vcs is not svn"
		response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
		return
	}

	// Distinguish event type according to the payload.
	switch payload.Event {
	case api.SVNWebhookCommit:
		log.Info("svn webhook receive commit event")

		// Check if is a commit to special url
		// In same repo of SVN, commit in other dir will also trap webhook.
		needCreateVerion, commitLog, err := isCommitToSpecialURL(payload.CommitID, service)
		if err != nil || false == needCreateVerion {
			webhookResponse.ErrorMessage = "ignore"
			response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
			return
		}

		// Create version config.
		version := &api.Version{
			ServiceID:        serviceID,
			Operation:        api.IntegrationOperation,
			Name:             "ci_" + payload.CommitID,
			Description:      commitLog,
			Commit:           payload.CommitID,
			YamlDeploy:       api.DeployWithYaml,
			YamlDeployStatus: api.DeployNoRun,
			SecurityCheck:    false,
		}

		// Create version.
		webhookCreateVersion(serviceID, version)

	default:
		log.Info("svn webhook receive unknow event")
	}

	webhookResponse.ErrorMessage = "ok"
	response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
}

// isCommitToSpecialURL gets if the commit is to a specific url.
func isCommitToSpecialURL(commitID string, service *api.Service) (bool, string, error) {
	args := []string{"log", service.Repository.URL, "-r", commitID, "--username", service.Repository.Username,
		"--password", service.Repository.Password, "--non-interactive", "--trust-server-cert",
		"--no-auth-cache"}
	output, err := executil.RunInDir("./", "svn", args...)
	log.Info(string(output))

	if err != nil {
		log.ErrorWithFields("Error when call isCommitToSpecialURL", log.Fields{"error": err})
		return false, string(output), err

	}

	return strings.Contains(string(output), "r"+commitID), string(output), nil
}

// webhookCreateVersion creates a creatversion event by webhook data.
func webhookCreateVersion(serviceID string, version *api.Version) error {
	// Find service info from DB.
	ds := store.NewStore()
	defer ds.Close()
	service, err := ds.FindServiceByID(serviceID)
	if err != nil {
		message := fmt.Sprintf("Unable to find service %v", version.ServiceID)
		log.ErrorWithFields(message, log.Fields{"user_id": service.UserID, "error": err})
		return fmt.Errorf("%s", message)
	}

	if "" == version.URL {
		version.URL = service.Repository.URL
	}
	version.Operator = api.WebhookOperator

	// To create a version, we must first make sure repository is healthy.
	if service.Repository.Status != api.RepositoryHealthy {
		message := fmt.Sprintf("Repository of service %s is not healthy, current status %s", service.Name, service.Repository.Status)
		log.ErrorWithFields(message, log.Fields{"user_id": service.UserID})
		return fmt.Errorf("%s", message)
	}

	// Request looks good, now fill up initial version status.
	version.CreateTime = time.Now()
	version.Status = api.VersionPending

	// Create a new version in database. Note the version is NOT the final version:
	// there can be error when running tests or building docker image. The version
	// ID is only a record that a version build has occurred. If the version build
	// succeeds, it'll be added to the service and is considered as a final version;
	// otherwise, it is just a version recorded in database.
	_, err = ds.NewVersionDocument(version)
	if err != nil {
		message := "Unable to create version document in database"
		log.ErrorWithFields(message, log.Fields{"user_id": service.UserID, "error": err})
		return fmt.Errorf("%s", message)
	}

	// Start building the version asynchronously, and make sure event is successfully
	// created before return.
	err = sendCreateVersionEvent(service, version)
	if err != nil {
		message := "Unable to create build version job"
		log.ErrorWithFields(message, log.Fields{"user_id": service.UserID, "service": service, "version": version, "error": err})
		return fmt.Errorf("%s", message)
	}

	return nil
}

// webhookGitLab handler webhook data from gitlab.
//
// POST: /api/v0.1/{service_id}/webhook_gitlab
//
// RESPONSE: (WebhookResponse)
//  {
//    "error_msg": (string) set IFF the request fails.
//  }
func webhookGitLab(request *restful.Request, response *restful.Response) {
	// Read payload.
	payload := api.WebhookGitlab{}
	err := request.ReadEntity(&payload)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	var webhookResponse api.WebhookResponse
	serviceID := request.PathParameter("service_id")

	// Check vcs of service is gitlab.
	if api.Git != getServiceVcs(serviceID) {
		webhookResponse.ErrorMessage = "vcs is not git"
		response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
		return
	}

	// Distinguish event type according to the payload.
	eventType := payload["object_kind"].(string)

	// Ignore repeated events.
	if isNeedIgnoreGitlab(eventType, payload) {
		log.Info("Ignore repeated events.")
		webhookResponse.ErrorMessage = "ignore"
		response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
		return
	}

	// Set create version operation:
	// 	commit: integration
	// 	tag: integration and pblish
	// 	pull request: integration
	var operation api.VersionOperation
	if isReleaseTagGitlab(eventType, payload) {
		log.Info("It's a release tag event")
		operation = api.IntegrationOperation + api.PublishOperation + api.DeployOperation
	} else {
		operation = api.IntegrationOperation
	}

	// Create version config.
	version := createVersionGitlabConfig(serviceID, eventType, payload, operation)
	if nil == version {
		webhookResponse.ErrorMessage = "unknow"
		response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
		return
	}

	// Create version.
	errCreateVersion := webhookCreateVersion(serviceID, version)
	if nil != errCreateVersion {
		webhookResponse.ErrorMessage = errCreateVersion.Error()
		response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
		return
	}

	webhookResponse.ErrorMessage = "ok"
	response.WriteHeaderAndEntity(http.StatusOK, webhookResponse)
}

// isNeedIgnoreGitlab gets if the event needs to be ignored.
func isNeedIgnoreGitlab(eventType string, payload api.WebhookGitlab) bool {
	if api.GitlabWebhookPush == eventType {
		if payload["checkout_sha"] == nil {
			return true
		}
	} else if api.GitlabWebhookPullRequest == eventType {
		attributes := payload["object_attributes"].(map[string]interface{})
		if attributes["state"] != api.PRActionOpened {
			return true
		}
	} else if api.GitlabWebhookRelease == eventType {
		// Create a version in UI, it will create a tag in repo automatically.
		// Need to ignore it.
		if isAutoCreateReleaseTagGitlab(payload) {
			return true
		}
		if payload["checkout_sha"] == nil {
			return true
		}
	}
	return false
}

// isAutoCreateReleaseTagGitlab gets if the event is a auto create release
// tag event according to the payload.
func isAutoCreateReleaseTagGitlab(payload api.WebhookGitlab) bool {
	if nil != payload[api.GitlabWebhookFlagRef] &&
		strings.Contains(payload[api.GitlabWebhookFlagRef].(string), api.GitlabWebhookFlagTags) &&
		strings.HasSuffix(payload[api.GitlabWebhookFlagRef].(string), api.AutoCreateTagFlag) {
		return true
	}
	return false
}

func isReleaseTagGitlab(eventType string, payload api.WebhookGitlab) bool {
	if api.GitlabWebhookRelease == eventType {
		return true
	}
	return false
}

// createVersionGithubConfig creates a version config.
func createVersionGitlabConfig(serviceID string, eventType string, payload api.WebhookGitlab,
	operation api.VersionOperation) *api.Version {
	version := &api.Version{
		ServiceID:        serviceID,
		Operation:        operation,
		YamlDeploy:       api.DeployWithYaml,
		YamlDeployStatus: api.DeployNoRun,
		SecurityCheck:    false,
	}

	switch eventType {
	case api.GitlabWebhookPush:
		version.Name, version.Description, version.Commit = generateVersionFromGitlabPushData(payload)
		log.Info("webhook event: push")

	case api.GitlabWebhookPullRequest:
		version.Name, version.Description, version.URL, version.Commit = generateVersionFromGitlabPRData(payload)
		log.Info("webhook event: merge_request")

	case api.GitlabWebhookRelease:
		version.Name, version.Description, version.Commit = generateVersionFromGitlabRelData(payload)
		log.Infof("webhook event: release")

	default:
		log.Info("receive undefine webhook event")
		return nil
	}

	return version
}

// generateVersionFromPushData generates Version config from payload data.
// name:
//   tag: tag_commitId
//   commit: ci_commitId
// description: message commitId
func generateVersionFromGitlabPushData(payload api.WebhookGitlab) (name string, description string,
	commitID string) {
	// generate name
	namePrefix := "ci_"

	if nil != payload["checkout_sha"] {
		name = namePrefix + payload["checkout_sha"].(string)
	} else {
		name = namePrefix + uuid.NewV4().String()
	}

	// generate description
	if nil != payload["ref"] {
		description = payload["ref"].(string)
	} else {
		description = ""
	}
	if nil != payload["checkout_sha"] {
		commitID = payload["checkout_sha"].(string)
		description = description + "\r\n" + payload["checkout_sha"].(string)
	}

	log.Infof("webhook push event: name[%s] description[%s] commit[%s]", name, description, commitID)
	return name, description, commitID
}

// generateVersionFromPRData generates Version config from payload data.
// name: prNo_commitId
// description: title body commitId pr_url
func generateVersionFromGitlabPRData(payload api.WebhookGitlab) (name string, description string,
	url string, commitID string) {
	attributes := payload["object_attributes"].(map[string]interface{})
	lastCommit := attributes["last_commit"].(map[string]interface{})
	source := attributes["source"].(map[string]interface{})

	url = source["git_http_url"].(string)

	// analysis pr commit
	commitID = lastCommit["id"].(string)

	// generate name
	if nil != attributes["id"] {
		name = fmt.Sprintf("pr_%v_%s", attributes["id"], commitID)
	} else {
		name = fmt.Sprintf("pr_%s", commitID)
	}

	// generate description
	if nil != attributes["title"] {
		description = attributes["title"].(string)
	} else {
		description = ""
	}
	if nil != attributes["description"] && "" != attributes["description"] {
		description = description + "\r\n" + attributes["description"].(string)
	}
	description = description + "\r\n" + commitID + "\r\n" + url

	log.Infof("webhook pr event: name[%s] description[%s] url[%s] commit[%s]", name, description,
		url, commitID)
	return name, description, url, commitID
}

// generateVersionFromGitlabRelData generates Version config from payload data.
func generateVersionFromGitlabRelData(payload api.WebhookGitlab) (name string, description string,
	commitID string) {
	ref := payload[api.GitlabWebhookFlagRef].(string)

	// generate name
	var namePrefix string
	var tagName string

	namePrefix = "tag_"
	// e.g. refs/tags/v0.12
	tagName = strings.SplitN(ref, "/", -1)[2]

	if nil != payload["checkout_sha"] {
		if "" != tagName && canBeUsedInImageTag(tagName) {
			name = namePrefix + tagName
		} else {
			name = namePrefix + payload["checkout_sha"].(string)
		}
	} else {
		name = namePrefix + uuid.NewV4().String()
	}

	// generate description
	if nil != payload["message"] {
		description = payload["message"].(string)
	} else {
		description = ""
	}
	if nil != payload["checkout_sha"] {
		commitID = payload["checkout_sha"].(string)
		description = description + "\r\n" + commitID
	}

	log.Infof("webhook push event: name[%s] description[%s] commit[%s]", name, description, commitID)
	return name, description, commitID
}
