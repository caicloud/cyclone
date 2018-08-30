package mock

import (
	"github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"
)

const (
	NmClient  = "normal-client"
	ProClient = "project-client"
	LgClient  = "log-client"
)

type MockedClients struct {
	clients map[string]map[string]interface{}
}

func NewMockedClients() *MockedClients {
	return &MockedClients{
		clients: make(map[string]map[string]interface{}),
	}
}

func (c *MockedClients) Add(registry string, client interface{}, kind string) {
	_, ok := c.clients[registry]
	if !ok {
		c.clients[registry] = make(map[string]interface{})
	}
	c.clients[registry][kind] = client
}

func (c *MockedClients) GetClient(registry string) (*harbor.Client, error) {
	return nil, nil
}

func (c *MockedClients) GetProjectClient(registry string) (harbor.ProjectClienter, error) {
	if client, ok := c.clients[registry]; ok {
		return client[ProClient].(harbor.ProjectClienter), nil
	}
	return nil, errors.ErrorUnknownInternal.Error("no available client")
}

func (c *MockedClients) GetLogClient(registry string) (harbor.LogClienter, error) {
	if client, ok := c.clients[registry]; ok {
		return client[LgClient].(harbor.LogClienter), nil
	}
	return nil, errors.ErrorUnknownInternal.Error("no available client")
}

func (c *MockedClients) Delete(registry string) {
	delete(c.clients, registry)
}

func (c *MockedClients) Update(registry string, client *harbor.Client) {
	c.clients[registry][NmClient] = client
}

func (c *MockedClients) Refresh(rinfos []*models.RegistryInfo) {
}
