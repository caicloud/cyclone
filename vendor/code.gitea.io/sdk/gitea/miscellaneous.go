// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gitea

// ServerVersion returns the version of the server
func (c *Client) ServerVersion() (string, error) {
	var v = struct {
		Version string `json:"version"`
	}{}
	return v.Version, c.getParsedResponse("GET", "/version", nil, nil, &v)
}
