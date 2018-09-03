package harbor

import (
	"github.com/caicloud/cargo-admin/pkg/models"
)

type ClientManager interface {
	GetClient(registry string) (*Client, error)
	GetProjectClient(registry string) (ProjectClienter, error)
	GetLogClient(registry string) (LogClienter, error)
	Delete(registry string)
	Update(registry string, client *Client)
	Refresh(rinfos []*models.RegistryInfo)
}
