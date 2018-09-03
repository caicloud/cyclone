package cauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (c *CauthClient) GetUser(name string) (*User, error) {
	url := c.Host + UserPath(name)
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

	user := &User{}
	err = json.Unmarshal(b, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
