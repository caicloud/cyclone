package resource

import (
	"time"

	"github.com/caicloud/cargo-admin/pkg/env"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"
	"github.com/caicloud/cargo-admin/pkg/utils/domain"
)

const (
	DefaultRegAlias = "Default"
	DefaultRegName  = "default"

	initProjectName = "library"
)

func InitRegistry(conf *env.RegistrySpec) error {
	domain, err := domain.GetDomain(conf.Host)
	if err != nil {
		return err
	}

	_, err = harbor.LoginAndGetCookies(&harbor.Config{
		Host:     conf.Host,
		Username: conf.Username,
		Password: conf.Password,
	})
	if err != nil {
		return err
	}

	b, err := models.Registry.IsExistHost(conf.Host)
	if err != nil {
		return err
	}
	now := time.Now()
	if !b {
		err := models.Registry.Save(&models.RegistryInfo{
			Name:           DefaultRegName,
			Alias:          DefaultRegAlias,
			Host:           conf.Host,
			Domain:         domain,
			Username:       conf.Username,
			Password:       conf.Password,
			CreationTime:   now,
			LastUpdateTime: now,
		})
		if err != nil {
			return err
		}
		return nil
	}

	regInfo, err := models.Registry.FindByName(DefaultRegName)
	if err != nil {
		return err
	}

	if (regInfo.Username != conf.Username) || (regInfo.Password != conf.Password) {
		err := models.Registry.Update(DefaultRegName, DefaultRegAlias, conf.Username, conf.Password)
		if err != nil {
			return err
		}
	}

	return nil
}
