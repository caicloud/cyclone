package cauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// TODO(ChenDe): If no pagination (start, limit) set, cauth will return all roles, but this
// may be changed in the future.
func (c *CauthClient) ListRoles(subType SubjectType, subId string) (*RoleList, error) {
	url := c.Host + RolesPath(string(subType), subId)
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

	roles := &RoleList{}
	err = json.Unmarshal(b, roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
