package resource

import (
	"context"
	"math/rand"
	"time"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"
)

const (
	passwordLen = 12
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GetOrNewDockerAccount(ctx context.Context, registry, username string) (*types.DockerAccount, error) {
	reginfo, err := models.Registry.FindByName(registry)
	if err != nil {
		return nil, err
	}
	// 如果是 admin 用户，直接返回管理员账户即可
	if username == reginfo.Username {
		return &types.DockerAccount{
			Metadata: &types.DockerAccountMetadata{
				CreationTime: reginfo.CreationTime,
			},
			Spec: &types.DockerAccountSpec{
				Host:     reginfo.Domain,
				Username: reginfo.Username,
				Password: reginfo.Password,
			},
			Status: &types.DockerAccountStatus{},
		}, nil
	}

	dainfo, err := getOrNewDockerAccount(reginfo, username)
	if err != nil {
		log.Errorf("getOrNewDockerAccount eror %v", err)
		return nil, err
	}

	return &types.DockerAccount{
		Metadata: &types.DockerAccountMetadata{
			CreationTime: dainfo.CreateTime,
		},
		Spec: &types.DockerAccountSpec{
			Host:     reginfo.Domain,
			Username: dainfo.Username,
			Password: dainfo.Password,
		},
		Status: &types.DockerAccountStatus{},
	}, nil
}

func getOrNewDockerAccount(reginfo *models.RegistryInfo, username string) (*models.DockerAccountInfo, error) {
	exist, err := models.DockerAccount.IsExist(reginfo.Name, username)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	if !exist {
		// 如果不存在，就创建一个，并写到数据库中
		da := newDockerAccount(username, reginfo.Domain)
		err = models.DockerAccount.Save(&models.DockerAccountInfo{
			Username:   da.Username,
			Password:   da.Password,
			Registry:   reginfo.Name,
			CreateTime: time.Now(),
		})
		if err != nil {
			log.Errorf("mongo error: %v", err)
			return nil, ErrorUnknownInternal.Error(err)
		}
	}

	dainfo, err := models.DockerAccount.FindByName(reginfo.Name, username)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	return dainfo, nil
}

func newDockerAccount(username string, domain string) *types.DockerAccountSpec {
	return &types.DockerAccountSpec{
		Host:     domain,
		Username: username,
		Password: genStr(passwordLen),
	}
}

func genStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
