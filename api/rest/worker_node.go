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
	"strconv"
	"strings"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"
	"github.com/emicklei/go-restful"
	docker_client "github.com/fsouza/go-dockerclient"
)

const WORKER_NODE_DOCKER_MIN_VERSION = "1.10.0"

// createSystemWorkerNode creates a system worker node.
//
// PUT: /api/v0.1/system_worker_nodes
//
// POST: /api/v0.1/:uid/services/
//
// PAYLOAD (WorkerNode):
//   {
//     "name": (string) Name of the worker node
//     "ip": (string) IP of the worker node
//     "docker_host": (string) DockerHost of the worker node
//     "total_resource": {
//       "memory": (float64) The memory config
//       "cpu": (float64) The CPU config
//     }
//   }
//
// RESPONSE: (WorkerNodeCreateResponse)
//  {
//    "node_id": (string) ID of the worker node.
//    "error_msg": (string) set IFF the request fails.
//  }
func createSystemWorkerNode(request *restful.Request, response *restful.Response) {
	// Read out posted service information.
	workerNode := api.WorkerNode{}
	err := request.ReadEntity(&workerNode)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	var createResponse api.WorkerNodeCreateResponse

	ds := store.NewStore()
	defer ds.Close()

	// Check repeated docker host.
	nodes, errRepeat := ds.FindWorkerNodesByDockerHost(workerNode.DockerHost)
	if errRepeat != nil {
		message := fmt.Sprintf("Check repeat docker host error %v", errRepeat)
		log.Error(message)
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}
	if len(nodes) != 0 {
		message := fmt.Sprintf("Docker host exists")
		log.Error(message)
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	// Check if the docker daemon is valid.
	if vaild, err := IsValidDockerHost(workerNode.DockerHost); vaild == false {
		message := fmt.Sprintf("Docker host isn't valid: %v", err)
		log.Error(message)
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	workerNode.LeftResource = workerNode.TotalResource
	nodeID, err := ds.NewSystemWorkerNodeDocument(&workerNode)
	if err != nil {
		message := fmt.Sprintf("Create new worker node err: %v", err)
		log.Error(message)
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
	}

	createResponse.NodeID = nodeID
	response.WriteHeaderAndEntity(http.StatusOK, createResponse)
}

//  convertVersionFromStringToInt convert the version from string to int.
func convertVersionFromStringToInt(version string) (int, error) {
	versions := strings.Split(version, ".")
	main, err := strconv.Atoi(versions[0])
	if err != nil {
		return 0, err
	}

	middle, err := strconv.Atoi(versions[1])
	if err != nil {
		return 0, err
	}

	little, err := strconv.Atoi(versions[2])
	if err != nil {
		return 0, err
	}

	return main*10000 + middle*100 + little, nil
}

// checkDockerVersion check wether if the node docker version is above the WORKER_NODE_DOCKER_MIN_VERSION.
func checkDockerVersion(nodeDockerVersion string) bool {
	nodeDockerVersionNum, err := convertVersionFromStringToInt(nodeDockerVersion)
	if err != nil {
		log.Errorf("Node docker version format err: %s", nodeDockerVersion)
		return false
	}

	log.Infof("Node docker version num is %d", nodeDockerVersionNum)

	minDockerVersionNum, _ := convertVersionFromStringToInt(WORKER_NODE_DOCKER_MIN_VERSION)
	if nodeDockerVersionNum < minDockerVersionNum {
		return false
	}

	return true
}

// IsValidDockerHost checks if the docker host is valid.
func IsValidDockerHost(dockerHost string) (bool, error) {
	client, err := docker_client.NewClient(dockerHost)
	if err != nil {
		return false, err
	}

	_, err = client.ListContainers(docker_client.ListContainersOptions{})
	if err != nil {
		return false, err
	}

	envs, err := client.Version()
	if err != nil {
		return false, err
	}

	bVaildDockerVersion := false
	for _, env := range *envs {
		if strings.HasPrefix(env, "Version") {
			log.Infof("The docker version of adding worker node is %s", strings.Split(env, "=")[1])
			if checkDockerVersion(strings.Split(env, "=")[1]) {
				bVaildDockerVersion = true
			}
			break
		}
	}

	if !bVaildDockerVersion {
		return false, fmt.Errorf("The docker version is not v1.10.1")
	}

	return true, nil
}

// getSystemWorkerNode gets a system worker node by node id.
//
// GET: /api/v0.1/system_worker_nodes/:node_id
//
// RESPONSE: (WorkerNodeGetResponse)
//  {
//    "service": (object) api.WorkerNode object.
//    "error_msg": (string) set IFF the request fails.
//  }
func getSystemWorkerNode(request *restful.Request, response *restful.Response) {
	var getResponse api.WorkerNodeGetResponse

	nodeID := request.PathParameter("node_id")

	ds := store.NewStore()
	defer ds.Close()

	result, err := ds.FindWorkerNodeByID(nodeID)
	if err != nil {
		message := fmt.Sprintf("Unable to find worker node %v", nodeID)
		log.ErrorWithFields(message, log.Fields{"error": err})
		getResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusNotFound, message)
		return
	}
	getResponse.WorkerNode = *result

	response.WriteEntity(getResponse)
}

// listSystemWorkerNodes returns all system worker nodes.
//
// GET: /api/v0.1/system_worker_nodes
//
// RESPONSE: (WorkerNodesListResponse)
//  {
//    "WorkerNodes": (array) a list of api.Service object.
//    "error_msg": (string) set IFF the request fails.
//  }
func listSystemWorkerNodes(request *restful.Request, response *restful.Response) {
	ds := store.NewStore()
	defer ds.Close()

	var listResponse api.WorkerNodesListResponse
	result, err := ds.FindSystemWorkerNode()

	if err != nil {
		message := "Unable to list system worker nodes"
		log.ErrorWithFields(message, log.Fields{"error": err})
		listResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusNotFound, message)
		return
	}
	listResponse.WorkerNodes = result

	response.WriteEntity(listResponse)
}

// deleteSystemWorkerNode deletes the system worker node by node_id.
//
// DELETE: /api/v0.1/system_worker_nodes/:node_id
//
// RESPONSE: (WorkerNodeDelResponse)
//  {
//    "result": (string) the result of deleting node
//    "error_msg": (string) set IFF the request fails.
//  }
func deleteSystemWorkerNode(request *restful.Request, response *restful.Response) {
	nodeID := request.PathParameter("node_id")

	var deleteResponse api.WorkerNodeDelResponse
	ds := store.NewStore()
	defer ds.Close()

	// Delete the node in DB
	err := ds.DeleteWorkerNodeByID(nodeID)
	if err != nil {
		message := "Unable to delete worker node"
		log.ErrorWithFields(message, log.Fields{"node_id": nodeID, "error": err})
		deleteResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, message)
		return
	}

	deleteResponse.Result = "success"
	response.WriteEntity(deleteResponse)
}
