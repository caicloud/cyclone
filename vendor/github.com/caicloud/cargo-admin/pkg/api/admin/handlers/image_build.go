package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/docker"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"
	"github.com/caicloud/cargo-admin/pkg/resource"

	"github.com/caicloud/nirvana/errors"
	"github.com/caicloud/nirvana/log"

	docker_types "github.com/docker/docker/api/types"
	"gopkg.in/mgo.v2"
)

func ListDockerfiles(ctx context.Context, tenant string) (*types.ListResponse, error) {
	dockerfiles, err := resource.GetDockerfiles()
	if err != nil {
		return nil, ErrorUnknownInternal.Error("get dockerfiles error")
	}

	return types.NewListResponse(len(dockerfiles), dockerfiles), nil
}

func BuildImage(ctx context.Context, tenant, registry, project, tag string, data io.Reader) error {
	log.Infof("Start to build image from Dockerfile to %s/%s", registry, project)
	regInfo, err := models.Registry.FindByName(registry)
	if err != nil {
		if err == mgo.ErrNotFound {
			return errors.NotFound.Build(types.RegistryNotExist, "${registry} not exist").Error(registry)
		}
		log.Errorf("Get registry %s error", registry)
		return ErrorUnknownInternal.Error(fmt.Sprintf("Get registry %s error", registry))
	}
	exist, err := models.Project.IsExist(tenant, registry, project)
	if err != nil {
		msg := fmt.Sprintf("Failed to check existence of project %s", project)
		log.Error(msg)
		return ErrorUnknownInternal.Error(msg)
	}
	if !exist {
		log.Errorf("Project %s not found in registry %s for tenant %s", project, registry, tenant)
		return errors.NotFound.Build(types.ProjectNotExist, "${project} not exist").Error(project)
	}

	imageName := fmt.Sprintf("%s/%s/%s", regInfo.Domain, project, tag)
	log.Infof("image to be build is: %s", imageName)
	rsp, err := docker.Client.ImageBuild(context.Background(), data, docker_types.ImageBuildOptions{
		Tags: []string{imageName},
	})
	if err != nil {
		return ErrorUnknownInternal.Error(fmt.Sprintf("build image error: %v", err))
	}
	defer rsp.Body.Close()
	defer imagesRemove([]string{imageName})

	d, e := ioutil.ReadAll(rsp.Body)
	if e != nil || !docker.IsBuildSucceed(string(d)) {
		log.Errorf("build image error, response body: %s", string(d))
		return ErrorUnknownInternal.Error("build image error")
	}
	log.Infof("build image %s succeed", tag)

	authConfig := docker_types.AuthConfig{
		Username:      regInfo.Username,
		Password:      regInfo.Password,
		ServerAddress: regInfo.Domain,
	}
	loginResp, err := docker.Client.RegistryLogin(context.Background(), authConfig)
	if err != nil {
		log.Errorf("login to %s as %s error: %v", regInfo.Domain, regInfo.Username, err)
		return ErrorNotAllowed.Error("docker login failed")
	}
	log.Infof("login registry %s: %v", regInfo.Domain, loginResp)

	authConfigBytes, err := json.Marshal(&authConfig)
	authConfigB64 := base64.StdEncoding.EncodeToString(authConfigBytes)
	pushResp, err := docker.Client.ImagePush(context.Background(), imageName, docker_types.ImagePushOptions{
		RegistryAuth: authConfigB64,
	})
	if err != nil {
		log.Errorf("push %s error: %v", imageName, err)
		return ErrorUnknownInternal.Error(fmt.Sprintf("push %s error: %v", imageName, err))
	}

	// This must be called to finish the push request
	d, e = ioutil.ReadAll(pushResp)
	pushResp.Close()
	if e != nil || !docker.IsPushSucceed(string(d)) {
		log.Errorf("push %s error, response body: %s", imageName, string(d))
		return ErrorUnknownInternal.Error("push image error")
	}
	log.Infof("push %s succeed")

	return nil
}
