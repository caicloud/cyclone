/*
Copyright 2017 caicloud authors. All rights reserved.

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

package store

import (
	"fmt"

	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
	encryptutil "github.com/caicloud/cyclone/pkg/util/encrypt"
)

// encryptPasswordsForProjects encrypts passwords for projects before them are stored.
func encryptPasswordsForProjects(project *api.Project, saltKey string) error {
	fmt.Printf("key: %s; len: %d\n", saltKey, len(saltKey))
	// Encrypt the passwords.
	if project.Registry != nil {
		encryptPwd, err := encryptutil.Encrypt(project.Registry.Password, saltKey)
		if err != nil {
			log.Errorf("fail to encrypt registry password for project %s as %v", project.Name, err)
			return err
		}

		project.Registry.Password = encryptPwd
	}

	scm := project.SCM
	if scm != nil {
		// Tokens are used when SCM is Github or Gitlab or svn, and passwords are not stored.
		// So only need to encrypt tokens.
		if scm.Type == api.Github || scm.Type == api.Gitlab || scm.Type == api.SVN {
			encryptToken, err := encryptutil.Encrypt(scm.Token, saltKey)
			if err != nil {
				log.Errorf("fail to encrypt SCM token for project %s as %v", project.Name, err)
				return err
			}
			project.SCM.Token = encryptToken
		}
	}

	return nil
}

// decryptPasswordsForProjects decrypts passwords for projects after them are got from DB.
func decryptPasswordsForProjects(project *api.Project, saltKey string) error {
	// Decrypt the passwords.
	if project.Registry != nil {
		decryptPwd, err := encryptutil.Decrypt(project.Registry.Password, saltKey)
		if err != nil {
			log.Errorf("fail to decrypt registry password for project %s as %v", project.Name, err)
			return err
		}

		project.Registry.Password = decryptPwd
	}

	scm := project.SCM
	if scm != nil {
		// Tokens are used when SCM is Github or Gitlab or svn, and passwords are not stored.
		// So only need to decrypt tokens.
		if scm.Type == api.Github || scm.Type == api.Gitlab || scm.Type == api.SVN {
			decryptToken, err := encryptutil.Decrypt(scm.Token, saltKey)
			if err != nil {
				log.Errorf("fail to decrypt SCM token for project %s as %v", project.Name, err)
				return err
			}

			project.SCM.Token = decryptToken
		}
	}

	return nil
}
