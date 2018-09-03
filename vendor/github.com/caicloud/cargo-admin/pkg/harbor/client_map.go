package harbor

import (
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"
)

type ClientMap map[string]*Client

var ClientMgr ClientManager = ClientMap(make(map[string]*Client))

func (clients ClientMap) GetClient(registry string) (*Client, error) {
	_, ok := clients[registry]
	if !ok {
		rinfo, err := models.Registry.FindByName(registry)
		if err != nil {
			return nil, err
		}
		cli, err := newClient(&Config{Host: rinfo.Host, Username: rinfo.Username, Password: rinfo.Password})
		if err != nil {
			return nil, err
		}
		clients[registry] = cli
	}
	return clients[registry], nil
}

func (clients ClientMap) GetProjectClient(registry string) (ProjectClienter, error) {
	client, err := clients.GetClient(registry)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (clients ClientMap) GetLogClient(registry string) (LogClienter, error) {
	client, err := clients.GetClient(registry)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (clients ClientMap) Delete(registry string) {
	delete(clients, registry)
}

func (clients ClientMap) Update(registry string, client *Client) {
	clients[registry] = client
}

func (clients ClientMap) Refresh(rinfos []*models.RegistryInfo) {
	for _, rinfo := range rinfos {
		log.Infof("refresh harbor client, registry: %s, host: %s", rinfo.Name, rinfo.Host)
		cli, err := newClient(&Config{Host: rinfo.Host, Username: rinfo.Username, Password: rinfo.Password})
		if err != nil {
			log.Errorf("refresh harbor client error: %v", err)
			ClientMgr.Delete(rinfo.Name)
			continue
		}
		ClientMgr.Update(rinfo.Name, cli)
	}
}
