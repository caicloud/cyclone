package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/docker"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/errors"
	"github.com/caicloud/nirvana/log"

	docker_types "github.com/docker/docker/api/types"
	"gopkg.in/mgo.v2"
)

func UploadImage(ctx context.Context, tid, registry, project string, data io.Reader) (*types.ImageUploadStats, error) {
	log.Infof("Start to upload images by TAR to %s/%s", registry, project)
	regInfo, err := models.Registry.FindByName(registry)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.NotFound.Build(types.RegistryNotExist, "${registry} not exist").Error(registry)
		}
		log.Errorf("Get registry %s error", registry)
		return nil, ErrorUnknownInternal.Error(fmt.Sprintf("Get registry %s error", registry))
	}
	exist, err := models.Project.IsExist(tid, registry, project)
	if err != nil {
		msg := fmt.Sprintf("Failed to check existence of project %s", project)
		log.Error(msg)
		return nil, ErrorUnknownInternal.Error(msg)
	}
	if !exist {
		log.Errorf("Project %s not found in registry %s for tenant %s", project, registry, tid)
		return nil, errors.NotFound.Build(types.ProjectNotExist, "${project} not exist").Error(project)
	}

	result := &types.ImageUploadResult{
		Stats: &types.ImageUploadStats{
			Succeed: make([]string, 0),
			Failed:  make([]string, 0),
		},
	}

	rsp, err := docker.Client.ImageLoad(context.Background(), data, true)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(fmt.Sprintf("Load images error: %v", err))
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(fmt.Sprintf("Read response body error: %v", err))
	}
	if docker.IsBadTarFile(string(body)) {
		return nil, errors.BadRequest.Build(types.BadImageFile, "${msg}").Error("Not a valid image file")
	}

	loaded := docker.LoadedImages(string(body))
	if len(loaded) == 0 {
		return nil, errors.BadRequest.Build(types.BadImageFile, "${msg}").Error("No images found in the file")
	}

	log.Infof("Loaded images: %v", loaded)
	defer imagesRemove(loaded)

	images := &types.UploadImages{Images: make([]types.ImageItem, 0)}
	for _, img := range loaded {
		_, _, repo, tag, ok := docker.ParseName(img)
		if !ok {
			log.Errorf("Invalid image: %s", img)
			return nil, ErrorUnknownInternal.Error(fmt.Sprintf("Invalid tag: %s", img))
		}
		images.Add(types.ImageItem{
			Origin: img,
			Image:  fmt.Sprintf("%s/%s/%s:%s", regInfo.Domain, project, repo, tag),
		})
	}

	toPush := make([]string, 0)
	for _, img := range images.Images {
		if img.Origin != img.Image {
			err := docker.Client.ImageTag(context.Background(), img.Origin, img.Image)
			if err != nil {
				result.AddFailed(img.Image)
				log.Errorf("Tag %s to %s error: %v", img.Origin, img.Image, err)
				continue
			}
		}
		toPush = append(toPush, img.Image)
	}
	defer imagesRemove(toPush)

	authConfig := docker_types.AuthConfig{
		Username:      regInfo.Username,
		Password:      regInfo.Password,
		ServerAddress: regInfo.Domain,
	}
	loginResp, err := docker.Client.RegistryLogin(context.Background(), authConfig)
	if err != nil {
		msg := fmt.Sprintf("Login to %s as %s error: %v", regInfo.Domain, regInfo.Username, err)
		log.Error(msg)
		return nil, ErrorUnauthentication.Error(msg)
	}
	log.Infof("Login registry %s: %v", regInfo.Domain, loginResp)

	authConfigBytes, err := json.Marshal(&authConfig)
	authConfigB64 := base64.StdEncoding.EncodeToString(authConfigBytes)

	wg := &sync.WaitGroup{}
	for _, image := range toPush {
		wg.Add(1)
		go func(img string) {
			defer wg.Done()

			log.Infof("Push image %s to %s", img, regInfo.Domain)
			pushResp, err := docker.Client.ImagePush(context.Background(), img, docker_types.ImagePushOptions{
				RegistryAuth: authConfigB64,
			})
			if err != nil {
				result.AddFailed(img)
				log.Errorf("Push %s error: %v", img, err)
				return
			}

			// This must be called to finish the push request
			d, e := ioutil.ReadAll(pushResp)
			pushResp.Close()
			if e != nil || !docker.IsPushSucceed(string(d)) {
				result.AddFailed(img)
				log.Errorf("push %s error, response body: %s", img, string(d))
				return
			}
			log.Infof("Push %s succeed", img)
			result.AddSucceed(img)
		}(image)
	}
	wg.Wait()

	return result.Stats, nil
}

func imagesRemove(imgs []string) {
	for _, img := range imgs {
		r, e := docker.Client.ImageRemove(context.Background(), img, docker_types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if e != nil {
			log.Warningf("Remove %s error: %v", img, e)
		} else {
			log.Infof("Image %v removed", r)
		}
	}
}
