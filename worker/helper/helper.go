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

package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/wait"
	"github.com/caicloud/cyclone/utils"
	"github.com/caicloud/cyclone/worker/ci"
	"github.com/caicloud/cyclone/worker/ci/parser"
	"github.com/caicloud/cyclone/worker/ci/runner"
	"github.com/caicloud/cyclone/worker/ci/yaml"
	"github.com/caicloud/cyclone/worker/clair"
	steplog "github.com/caicloud/cyclone/worker/log"
	k8s_core_api "k8s.io/kubernetes/pkg/api"
	k8s_ext_api "k8s.io/kubernetes/pkg/apis/extensions"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	restclient "k8s.io/kubernetes/pkg/client/restclient"
)

// Application is the type for k8s deployment.
type Application struct {
	Deployment       k8s_ext_api.Deployment            `json:"deployment,omitempty"`
	Service          k8s_core_api.Service              `json:"service,omitempty"`
	DeleteOptions    k8s_core_api.DeleteOptions        `json:"deleteOptions,omitempty"`
	Pods             k8s_core_api.PodList              `json:"pods,omitempty"`
	Nodes            k8s_core_api.NodeList             `json:"nodes,omitempty"`
	PodEvents        map[string]k8s_core_api.EventList `json:"podEvents,omitempty"`
	DeploymentEvents k8s_core_api.EventList            `json:"deploymentEvents,omitempty"`
	ReplicaSetEvents k8s_core_api.EventList            `json:"replicaSetEvents,omitempty"`

	AutoProvisionPVC map[string]k8s_core_api.PersistentVolumeClaim `json:"autoProvisionPVC,omitempty"`
}

const (
	// KUBERNETES is the cluster type of kubernetes.
	KUBERNETES = "kubernetes"

	checkDeployStatusPeriod = 10 * time.Second
	checkDeployTimeout      = 10 * time.Minute
)

var (
	codeDeployReady = 1
)

// Result is the type for result.
type Result struct {
	position int
	err      error
}

// appVersionInfo defines which versions are used in an application.
type appVersionInfo struct {
	UserName      string                 `bson:"user_name,omitempty" json:"user_name,omitempty"`
	UserID        string                 `bson:"uid,omitempty" json:"uid,omitempty"`
	ServiceName   string                 `bson:"service_name,omitempty" json:"service_name,omitempty"`
	ClusterID     string                 `bson:"cid,omitempty" json:"cid,omitempty"`
	Namespace     string                 `bson:"namespace,omitempty" json:"namespace,omitempty"`
	Deployment    string                 `bson:"deployment,omitempty" json:"deployment,omitempty"`
	ContainerList []containerVersionInfo `bson:"containerlist,omitempty" json:"containerlist,omitempty"`
	DeployOk      bool                   `bson:"deploy_ok,omitempty" json:"deploy_ok,omitempty"`
	ImageName     string                 `bson:"image_name,omitempty" json:"image_name,omitempty"`
	ClusterType   string                 `bson:"cluster_type,omitempty" json:"cluster_type,omitempty"` // kubernetes, caicloud_claas, mesos
	ClusterHost   string                 `bson:"cluster_host,omitempty" json:"cluster_host,omitempty"`
	ClusterToken  string                 `bson:"cluster_token,omitempty" json:"cluster_token,omitempty"`
}

// containerVersionInfo defines which versions are used in an container.
type containerVersionInfo struct {
	ContainerName string `bson:"name,omitempty" json:"name,omitempty"`
	VersionName   string `bson:"version,omitempty" json:"version,omitempty"`
}

// ExecBuild exec the publish steps
// Step1: Prebuild
// Step2: Build
func ExecBuild(cmanager *ci.Manager, r *runner.Build) (err error) {
	// Run the prebuild step.
	if err = cmanager.ExecPreBuild(r); err != nil {
		return err
	}

	// Run the build step.
	if err = cmanager.ExecBuild(r); err != nil {
		return err
	}

	return nil

}

// ExecIntegration exec the integration steps
// Step1: set up service network
// Step2: integration
// Step3: tear down service network
func ExecIntegration(cmanager *ci.Manager, r *runner.Build) (err error) {
	// Set up the networks and volumes, then run the integration.
	if err = cmanager.Setup(r); err != nil {
		return err
	}

	defer func() {
		if err := cmanager.Teardown(r); err != nil {
			log.ErrorWithFields("Unable to tear down the build", log.Fields{"err": err})
		}
	}()

	// Integration
	if err = cmanager.ExecIntegration(r); err != nil {
		return err
	}

	return nil
}

// ExecPublish exec the publish steps
// Step1: Push images
// Step2: Post build
func ExecPublish(cmanager *ci.Manager, r *runner.Build) (err error) {
	//Push image.
	if err = cmanager.ExecPublish(r); err != nil {
		return err
	}

	if err = cmanager.ExecPostBuild(r); err != nil {
		return err
	}
	return nil

}

// ExecDeploy exec the deploy steps.
func ExecDeploy(event *api.Event, dmanager *docker.Manager,
	r *runner.Build, tree *parser.Tree) (err error) {

	// Deploy by yaml
	if err := DoYamlDeploy(r, tree, event, dmanager); err != nil {
		return err
	}

	// Deploy by deploy plans
	if err := DoPlansDeploy(r.IsPushImageSuccess(), event, dmanager); err != nil {
		return err
	}

	return nil
}

// DoYamlDeploy is a wrapper of deploy to do some extra work by yaml information.
func DoYamlDeploy(r *runner.Build, tree *parser.Tree, event *api.Event, dmanager *docker.Manager) error {
	if event.Version.YamlDeploy == api.NotDeployWithYaml || len(tree.DeployConfig.Applications) == 0 {
		log.Infof("Skip deploy due to deploy section not defined or yaml deploy not be choosed")
		return nil
	}

	if r.IsPushImageSuccess() == false {
		return fmt.Errorf("Failed to deploy version due to build section undefined or push image failed.")
	}

	imagename, ok := event.Data["image-name"]
	tagname, ok2 := event.Data["tag-name"]

	if !ok || !ok2 {
		return fmt.Errorf("Unable to retrieve image name")
	}
	imageName := imagename.(string) + ":" + tagname.(string)

	for _, application := range tree.DeployConfig.Applications {
		if err := updateContainerInClusterWithYaml(event.Service.UserID, imageName, application); err != nil {
			event.Version.YamlDeployStatus = api.DeployFailed
			return err
		}
	}
	event.Version.YamlDeployStatus = api.DeployPending
	return nil
}

// DoPlansDeploy is a wrapper of deploy to do some extra work by Plans information.
func DoPlansDeploy(bHasPublishSuccessful bool, event *api.Event, dmanager *docker.Manager) error {
	if bHasPublishSuccessful == false {
		return fmt.Errorf("Failed to deploy version due to build section undefined or push image failed.")
	}
	imagename, ok := event.Data["image-name"]
	tagname, ok2 := event.Data["tag-name"]

	if !ok || !ok2 {
		return fmt.Errorf("Unable to retrieve image name")
	}
	imageName := imagename.(string) + ":" + tagname.(string)

	for i := 0; i < len(event.Version.DeployPlansStatuses); i++ {
		plan := event.Version.DeployPlansStatuses[i]
		if err := updateContainerInClusterWithPlan(event.Service.UserID, imageName, plan.Config); err != nil {
			event.Version.DeployPlansStatuses[i].Status = api.DeployFailed
		}
		event.Version.DeployPlansStatuses[i].Status = api.DeployPending
	}
	return nil
}

// Publish is an async handler for build and push the image to
// private registry.
func Publish(event *api.Event, dmanager *docker.Manager) error {
	steplog.InsertStepLog(event, steplog.BuildImage, steplog.Start, nil)
	if err := dmanager.BuildImage(event, steplog.Output); err != nil {
		steplog.InsertStepLog(event, steplog.BuildImage, steplog.Stop, err)
		return err
	}
	steplog.InsertStepLog(event, steplog.BuildImage, steplog.Finish, nil)

	steplog.InsertStepLog(event, steplog.PushImage, steplog.Start, nil)
	if err := dmanager.PushImage(event, steplog.Output); err != nil {
		steplog.InsertStepLog(event, steplog.PushImage, steplog.Stop, err)
		return err
	}

	if err := clair.Analysis(event, dmanager); err != nil {
		log.ErrorWithFields("Unable to analysis by clair", log.Fields{"err": err})
	}
	steplog.InsertStepLog(event, steplog.PushImage, steplog.Finish, nil)
	return nil
}

// updateContainerInClusterWithYaml func use to update container in cluster according the caicloud.yaml.
func updateContainerInClusterWithYaml(userID, imageName string, application yaml.Application) error {
	clusterName := application.ClusterName
	namespaceName := application.NamespaceName
	deploymentName := application.DeploymentName
	for _, containerName := range application.Containers {
		log.InfoWithFields("Send post request to updateImage API for yaml deploy: ",
			log.Fields{
				"user id":     userID,
				"cluster":     clusterName,
				"partition":   namespaceName,
				"application": deploymentName,
				"container":   containerName,
				"image":       imageName,
			})
		if application.ClusterType == KUBERNETES {
			if err := InvokeUpdateImageK8sAPI(deploymentName, namespaceName, containerName, imageName,
				application.ClusterHost, application.ClusterToken); err != nil {
				log.ErrorWithFields("Failed to deploy with yaml information use k8s api", log.Fields{"err": err})
				return err
			}
		} else {
			consoleWebEndpoint := osutil.GetStringEnv("CONSOLE_WEB_ENDPOINT", "http://127.0.0.1:3000")
			endpoint := consoleWebEndpoint + "/api/application/updateImage"
			if err := InvokeUpdateImageAPI(userID, deploymentName, clusterName, namespaceName,
				containerName, imageName, endpoint); err != nil {
				log.ErrorWithFields("Failed to deploy with yaml information", log.Fields{"err": err})
				return err
			}
		}
	}
	return nil
}

// updateContainerInClusterWithPlan func use to update container in cluster according the plan setting.
func updateContainerInClusterWithPlan(userID, imageName string, application api.DeployConfig) error {
	consoleWebEndpoint := osutil.GetStringEnv("CONSOLE_WEB_ENDPOINT", "http://127.0.0.1:3000")
	endpoint := consoleWebEndpoint + "/api/application/updateImage"

	clusterName := application.ClusterID
	namespaceName := application.Namespace
	deploymentName := application.Deployment
	for _, containerName := range application.Containers {
		// Web may send the empty contianter name, so there need make some judgment
		if containerName == "" {
			continue
		}
		log.InfoWithFields("Send post request to updateImage API for plan deploy: ",
			log.Fields{
				"user id":     userID,
				"cluster":     clusterName,
				"partition":   namespaceName,
				"application": deploymentName,
				"container":   containerName,
				"image":       imageName,
			})
		if err := InvokeUpdateImageAPI(userID, deploymentName, clusterName, namespaceName,
			containerName, imageName, endpoint); err != nil {
			log.ErrorWithFields("Failed to deploy with plan information", log.Fields{"err": err})
			return err
		}
	}
	return nil
}

// ExecDeployCheck keeps call console-web API to check deploy status
// of the version util timeout or error occurred or success. It will write finalStatus to DB.
func ExecDeployCheck(event *api.Event, tree *parser.Tree) {
	// yaml deploy state check
	DoYamlDeployCheck(event, tree)

	// plan deploy state check
	DoPlanDeployCheck(event)
}

// DoYamlDeployCheck uses for yaml deploy state check.
func DoYamlDeployCheck(event *api.Event, tree *parser.Tree) {
	appList := []appVersionInfo{}
	value := []int{}
	if event.Version.YamlDeployStatus == api.DeployPending {
		finalStatus := api.DeploySuccess
		for _, app := range tree.DeployConfig.Applications {
			appList = append(appList, *parseAppVersionInfo(event, &app))
		}

		log.InfoWithFields("About to check deploy state for yaml depoly", log.Fields{
			"version_id": event.Version.VersionID,
			"applist":    appList,
		})

		if false == checkDeployStatus(event, appList, &value) {
			finalStatus = api.DeployFailed
		}

		event.Version.YamlDeployStatus = finalStatus
	}
}

// DoPlanDeployCheck uses for plan deploy state check.
func DoPlanDeployCheck(event *api.Event) {
	// Plan deploy check
	appList := []appVersionInfo{}
	value := []int{}
	for i, plan := range event.Version.DeployPlansStatuses {
		// Only check pending status
		if plan.Status != api.DeployPending {
			continue
		}
		event.Version.DeployPlansStatuses[i].Status = api.DeploySuccess
		appList = append(appList, *getAppVersionInfoFromPlan(event, &(plan.Config)))
	}
	log.InfoWithFields("About to check deploy state for plan depoly", log.Fields{
		"version_id": event.Version.VersionID,
		"applist":    appList,
	})

	checkDeployStatus(event, appList, &value)

	for _, position := range value {
		if position != -1 {
			event.Version.DeployPlansStatuses[position].Status = api.DeployFailed
		}
	}
}

// checkOneDeployStatus func use to check deploy update status once.
func checkOneDeployStatus(versionID string, checkChan chan Result, app appVersionInfo, position int) {
	err := wait.Poll(checkDeployStatusPeriod, checkDeployTimeout, func() (bool, error) {
		var err error
		if app.ClusterType == "kubernetes" {
			err = InvokeCheckDeployStateK8sAPI(app)
		} else {
			consoleWebEndpoint := osutil.GetStringEnv("CONSOLE_WEB_ENDPOINT", "http://127.0.0.1:3000")
			endpoint := consoleWebEndpoint + "/api/application/checkVersionDeployState"
			err = InvokeCheckDeployStateAPI(app, endpoint)
		}

		if err != nil {
			// May Failed because of network problem, just print error
			log.ErrorWithFields("Failed to call checkDeployAPI", log.Fields{
				"applicationName": app.Deployment,
				"err":             err,
			})
			return false, err
		}
		return true, nil
	})
	if err != nil {
		// just print error
		log.ErrorWithFields("Failed to check deploy status", log.Fields{
			"version_id": versionID,
			"err":        err,
		})
	}
	checkChan <- Result{position, err}
}

// checkDeployStatus func use to check deploy update status.
func checkDeployStatus(event *api.Event, appList []appVersionInfo, value *[]int) bool {
	// In e2e-test, we dont really send a http request. The call will be always successful.
	if event.Service.UserID == utils.DeployUID {
		return true
	}

	checkChan := make(chan Result, len(appList))
	exitChan := make(chan bool, 1)
	defer func() {
		close(checkChan)
		close(exitChan)
	}()

	for i, app := range appList {
		go checkOneDeployStatus(event.Version.VersionID, checkChan, app, i)
	}

	go checkUpdateResult(len(appList), checkChan, exitChan, value)

	return <-exitChan
}

// checkUpdateResult func collect all the check result.
func checkUpdateResult(length int, checkChan chan Result, exitChan chan bool, value *[]int) {
	final := true
	for i := 0; i < length; i++ {
		result := <-checkChan
		if result.err != nil {
			final = false
			*value = append(*value, result.position)
		}
	}
	exitChan <- final
}

// parseAppVersionInfo func parse the yaml's application struct to appVersionInfo struct.
func parseAppVersionInfo(event *api.Event, a *yaml.Application) *appVersionInfo {
	containerList := []containerVersionInfo{}

	for _, c := range a.Containers {
		containerList = append(containerList, containerVersionInfo{
			ContainerName: c,
			VersionName:   event.Version.Name,
		})
	}

	imagename, _ := event.Data["image-name"]
	tagname, _ := event.Data["tag-name"]
	imageName := imagename.(string) + ":" + tagname.(string)

	return &appVersionInfo{
		UserName:    event.Service.Username,
		UserID:      event.Service.UserID,
		ServiceName: event.Service.Name,
		// Actually this is a cluster id now..
		ClusterID:     a.ClusterName,
		Namespace:     a.NamespaceName,
		Deployment:    a.DeploymentName,
		ContainerList: containerList,
		DeployOk:      false,
		ImageName:     imageName,
		ClusterType:   a.ClusterType,
		ClusterHost:   a.ClusterHost,
		ClusterToken:  a.ClusterToken,
	}
}

// getAppVersionInfoFromPlan func parse the plan's application struct to appVersionInfo struct.
func getAppVersionInfoFromPlan(event *api.Event, plan *api.DeployConfig) *appVersionInfo {
	containerList := []containerVersionInfo{}

	for _, c := range plan.Containers {
		// Web may send the empty contianter name, so there need make some judgment
		if c == "" {
			continue
		}
		containerList = append(containerList, containerVersionInfo{
			ContainerName: c,
			VersionName:   event.Version.Name,
		})
	}

	imagename, _ := event.Data["image-name"]
	tagname, _ := event.Data["tag-name"]
	imageName := imagename.(string) + ":" + tagname.(string)

	return &appVersionInfo{
		UserName:    event.Service.Username,
		UserID:      event.Service.UserID,
		ServiceName: event.Service.Name,
		// Actually this is a cluster id now..
		ClusterID:     plan.ClusterID,
		Namespace:     plan.Namespace,
		Deployment:    plan.Deployment,
		ContainerList: containerList,
		DeployOk:      false,
		ImageName:     imageName,
	}
}

// InvokeUpdateImageAPI used for call api to updape application in cluster.
func InvokeUpdateImageAPI(userID, applicationName, clusterName, partitionName,
	containerName, imageName, endpoint string) error {
	// In e2e-test, we dont really send a http request. The call will be always successful.
	if userID == utils.DeployUID {
		return nil
	}
	// Set up form data.
	values := make(url.Values)
	values.Set("uid", userID)
	values.Set("cid", clusterName)
	values.Set("partitionName", partitionName)
	values.Set("applicationName", applicationName)
	values.Set("containerName", containerName)
	values.Set("imageName", imageName)

	// Build a client.
	client := &http.Client{}
	// Submit form.
	resp, err := client.PostForm(endpoint, values)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.DebugWithFields("invokeUpdateImageAPI response",
		log.Fields{
			"status code": resp.StatusCode,
			"body":        string(body[:]),
		})
	// Check response.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %v, response body: %v", resp.StatusCode, string(body[:]))
	}

	return nil
}

// InvokeCheckDeployStateAPI func call caicloud api to check whether the updating is successful.
func InvokeCheckDeployStateAPI(app appVersionInfo, endpoint string) error {
	jsonStr, err := json.Marshal(app)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json;charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.InfoWithFields("InvokeCheckDeployStateAPI response",
		log.Fields{
			"status code": resp.StatusCode,
			"body":        string(body[:]),
		})

	// Check status code.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %v, response body: %v", resp.StatusCode, string(body[:]))
	}

	result := make(map[string]int)
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	code, ok := result["code"]
	if !ok || code != codeDeployReady {
		return fmt.Errorf("error")
	}

	return nil
}

// InvokeUpdateImageK8sAPI invokes k8s api to update the depolyment.
func InvokeUpdateImageK8sAPI(deploymentName, namespaceName, containerName,
	imageName, host, token string) error {
	k8sClient, err := NewClientWithToken(host, token)
	if err != nil {
		return err
	}

	// Get applciton depolyment
	var deployment *k8s_ext_api.Deployment
	if deployment, err = k8sClient.Deployments(namespaceName).Get(deploymentName); err != nil {
		return err
	}

	log.Infof("deployment %+v", *deployment)
	for i := 0; i < len(deployment.Spec.Template.Spec.Containers); i++ {
		if deployment.Spec.Template.Spec.Containers[i].Name == containerName {
			deployment.Spec.Template.Spec.Containers[i].Image = imageName
		}
	}

	// Update deployment. Skip updating deployment if name is empty.
	if deployment.Name != "" {
		if _, err := k8sClient.Deployments(namespaceName).Update(deployment); err != nil {
			return err
		}
	}

	return nil
}

// InvokeCheckDeployStateK8sAPI func call k8s api to check whether the updating is successful.
func InvokeCheckDeployStateK8sAPI(app appVersionInfo) error {
	k8sClient, err := NewClientWithToken(app.ClusterHost, app.ClusterToken)
	if err != nil {
		return err
	}

	// Get applciton depolyment
	var deployment *k8s_ext_api.Deployment
	if deployment, err = k8sClient.Deployments(app.Namespace).Get(app.Deployment); err != nil {
		return err
	}

	log.Infof("deployment %+v", *deployment)
	for i := 0; i < len(deployment.Spec.Template.Spec.Containers); i++ {
		if checkInArray(app.ContainerList, deployment.Spec.Template.Spec.Containers[i].Name) &&
			deployment.Spec.Template.Spec.Containers[i].Image != app.ImageName {
			return fmt.Errorf("update fail, %s vs %s", deployment.Spec.Template.Spec.Containers[i].Image, app.ImageName)
		}
	}
	return nil
}

// checkInArray func use to check  whether contianerName is in containerList.
func checkInArray(containerList []containerVersionInfo, contianerName string) bool {
	for _, container := range containerList {
		if container.ContainerName == contianerName {
			return true
		}
	}

	return false
}

// NewClientWithToken creates a new kubernetes client using BearerToken.
func NewClientWithToken(host, token string) (*clientset.Clientset, error) {
	config := &restclient.Config{
		Host:        host,
		BearerToken: token,
		Insecure:    true,
	}
	return clientset.NewForConfig(config)
}
