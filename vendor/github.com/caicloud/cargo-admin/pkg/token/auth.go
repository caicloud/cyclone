package token

import (
	"strings"

	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"

	"gopkg.in/mgo.v2"
)

const CYCLONE_USER_PREFIX = "__cyclone__"

type basicAuth struct {
	Username string
	Password string
	Scope    []string
}

func (a *basicAuth) GetUsername() string {
	return a.Username
}

func (a *basicAuth) GetCycloneUserName() string {
	return strings.TrimPrefix(a.Username, CYCLONE_USER_PREFIX)
}

func (a *basicAuth) IsRegistryAdmin(rInfo *models.RegistryInfo) bool {
	return a.Username == rInfo.Username && a.Password == rInfo.Password
}

func (a *basicAuth) IsCycloneUser() bool {
	return strings.HasPrefix(a.Username, CYCLONE_USER_PREFIX)
}

func (a *basicAuth) ValidateUser(reginfo *models.RegistryInfo) error {
	if a.Username == reginfo.Username && a.Password == reginfo.Password {
		return nil
	}

	account, err := models.DockerAccount.FindByName(reginfo.Name, a.Username)
	if err != nil {
		if err == mgo.ErrNotFound {
			return ErrorUnauthentication.Error()
		}
		return ErrorUnknownInternal.Error(err)
	}
	if a.Password != account.Password {
		log.Errorf("password incorrect, username: %s", a.Username)
		return ErrorUnauthentication.Error()
	}

	return nil
}
