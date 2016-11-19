/*
Copyright 2016 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package docker

import (
	"errors"
)

// AuthConfig contains the username and password to access docker registry.
type AuthConfig struct {
	Username string
	Password string
}

// NewAuthConfig returns a new AuthConfig or returns an error.
func NewAuthConfig(username, password string) (*AuthConfig, error) {
	if username == "" || password == "" {
		return nil, errors.New("The username or password for docker registry is not set.")
	}
	return &AuthConfig{
		Username: username,
		Password: password,
	}, nil
}
