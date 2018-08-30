package cauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// TODO(ChenDe): If no pagination (start, limit) set, cauth will return all tenants, but this
// may be changed in the future.
func (c *CauthClient) ListTenants(username string) (*TenantList, error) {
	url := c.Host + TenantsPath(username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%s", string(b))
	}

	tenants := &TenantList{}
	err = json.Unmarshal(b, tenants)
	if err != nil {
		return nil, err
	}
	return tenants, nil
}
