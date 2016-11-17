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

package version

import (
	"net/http"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/wait"
	. "github.com/caicloud/cyclone/tests/common"
	gwebsocket "golang.org/x/net/websocket"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	DefaultServiceName = "test-basic-rest-version"
	DefaultVersionName = "v0.1.0"
	SecondVersionName  = "v0.2.0"
	DefaultTestRepo    = "https://github.com/caicloud/toy-dockerfile"
)

var _ = Describe("Version", func() {
	var (
		dockerManager       *docker.Manager
		service             *api.Service
		serviceResponse     *api.ServiceCreationResponse
		service_svn         *api.Service
		serviceResponse_svn *api.ServiceCreationResponse
		ws                  *gwebsocket.Conn
	)

	// Set up the serviceCM, versionCM and dockerManager and create a service.
	BeforeSuite(func() {
		// Wait cyclone to start.
		WaitComponents()

		// Get docker host and cert path.
		endpoint := osutil.GetStringEnv("DOCKER_HOST", DefaultDockerHost)
		certPath := osutil.GetStringEnv("DOCKER_CERT_PATH", "")

		// Get the username and password to access the docker registry.
		registryLocation := osutil.GetStringEnv("REGISTRY_LOCATION", DefaultRegistryAddress)
		registryUsername := osutil.GetStringEnv("REGISTRY_USERNAME", AdminUser)
		registryPassword := osutil.GetStringEnv("REGISTRY_PASSWORD", AdminPassword)
		registry := api.RegistryCompose{
			RegistryLocation: registryLocation,
			RegistryUsername: registryUsername,
			RegistryPassword: registryPassword,
		}

		// Create docker manager.
		var err error
		dockerManager, err = docker.NewManager(
			endpoint, certPath, registry)
		if err != nil {
			log.Fatalf("Unable to connect to docker daemon: %v", err)
		}

		// Create a test service.
		service = &api.Service{
			Name:        DefaultServiceName,
			Username:    AliceUser,
			Description: "A service just for test",
			Repository: api.ServiceRepository{
				URL: DefaultTestRepo,
				Vcs: api.Git,
			},
		}
		serviceResponse = &api.ServiceCreationResponse{}
		err = CreateService(AliceUID, service, serviceResponse)
		if err != nil {
			log.Errorf("unexpected error while creating service for user %v: %v", AliceUID, err)
		}

		// Wait up to 30 seconds until the service is successfully created.
		err = wait.Poll(2*time.Second, 30*time.Second, func() (bool, error) {
			serviceGetResponse := &api.ServiceGetResponse{}
			err := GetService(AliceUID, serviceResponse.ServiceID, serviceGetResponse)
			return serviceGetResponse.Service.Repository.Status == api.RepositoryHealthy, err
		})
		if err != nil {
			log.ErrorWithFields("unexpected repository status", log.Fields{"repository": service.Repository,
				"status": service.Repository.Status})
		}

		service_svn = &api.Service{
			Name:        DefaultServiceName + "_svn",
			Username:    AliceUser,
			Description: "A service just for testing svn",
			Repository: api.ServiceRepository{
				URL: DefaultTestRepo + "/trunk",
				Vcs: api.Svn,
			},
		}
		serviceResponse_svn = &api.ServiceCreationResponse{}
		err = CreateService(AliceUID, service_svn, serviceResponse_svn)
		if err != nil {
			log.Errorf("unexpected error while creating service for user %v: %v", AliceUID, err)
			return
		}

		// Wait up to 30 seconds until the service is successfully created.
		err = wait.Poll(2*time.Second, 30*time.Second, func() (bool, error) {
			serviceGetResponse := &api.ServiceGetResponse{}
			err := GetService(AliceUID, serviceResponse_svn.ServiceID, serviceGetResponse)
			return serviceGetResponse.Service.Repository.Status == api.RepositoryHealthy, err
		})
		if err != nil {
			log.ErrorWithFields("unexpected repository status", log.Fields{"repository": service_svn.Repository,
				"status": service_svn.Repository.Status})
			return
		}

		//create a websocket client
		ws, err = DialLogServer()
		if err != nil {
			log.Errorf("dail log server error: %v", err)
		}
	})

	Context("with right service id", func() {
		Context("with right version information", func() {
			var (
				versionResponse *api.VersionCreationResponse
			)

			It("should create version successfully.", func() {
				version := &api.Version{
					Name:        DefaultVersionName,
					Description: "A version just for test",
					ServiceID:   serviceResponse.ServiceID,
					Operation:   api.PublishOperation,
				}

				imageName := dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName, DefaultVersionName)
				// Forcibly remove image.
				dockerManager.RemoveImage(imageName)
				// Create the version.
				versionResponse = &api.VersionCreationResponse{}

				Expect(CreateVersion(AliceUID, version, versionResponse)).To(BeNil())
				Expect(versionResponse.ErrorMessage).To(Equal(""))
			})

			It("should be able to get version via HTTP GET method.", func() {
				versionGetResponse := &api.VersionGetResponse{}
				// Wait up to 120 seconds until docker image has been pushed.
				err := wait.Poll(3*time.Second, 120*time.Second, func() (bool, error) {
					err := GetVersion(AliceUID, versionResponse.VersionID, versionGetResponse)
					return versionGetResponse.Version.Status != api.VersionPending && versionGetResponse.Version.Status != api.VersionRunning, err
				})
				Expect(err).To(BeNil())
				Expect(versionGetResponse.ErrorMessage).To(Equal(""))
				Expect(versionGetResponse.Version).NotTo(BeNil())
				Expect(versionGetResponse.Version.Status).To(Equal(api.VersionHealthy))
				Expect(versionGetResponse.Version.YamlDeployStatus).To(Equal(api.DeployNoRun))
			})

			It("should clean up docker images from build process.", func() {
				imageName := dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName, DefaultVersionName)
				ok, err := dockerManager.IsImagePresent(imageName)
				Expect(ok).To(Equal(false))
				Expect(err).To(BeNil())
			})

			It("should pull newly created docker image successfully.", func() {
				imageName := dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName, DefaultVersionName)
				Expect(dockerManager.PullImage(imageName)).To(BeNil())
				ok, err := dockerManager.IsImagePresent(imageName)
				Expect(ok).To(Equal(true))
				Expect(err).To(BeNil())
			})

			// TODO: multi client GET the log.
			It("should push the log", func() {
				err := WatchLog(ws, APICreateVersion, AliceUID,
					serviceResponse.ServiceID, versionResponse.VersionID)
				Expect(err).To(BeNil())
			})

			It("should be able to list version via HTTP GET method.", func() {
				versionListResponse := &api.VersionListResponse{}
				Expect(ListVersions(AliceUID, serviceResponse.ServiceID, versionListResponse)).To(BeNil())
				Expect(versionListResponse.Versions).To(HaveLen(1))
				Expect(versionListResponse.Versions[0].Name).To(Equal(DefaultVersionName))
				Expect(versionListResponse.Versions[0].Status).To(Equal(api.VersionHealthy))
			})

			It("should add the new version to service.", func() {
				// Retrieve the service to verify the version has been added to the service.
				serviceGetResponse := &api.ServiceGetResponse{}
				Expect(GetService(AliceUID, serviceResponse.ServiceID, serviceGetResponse)).To(BeNil())
				Expect(serviceResponse.ErrorMessage).To(Equal(""))
				Expect(serviceGetResponse.Service.Repository.Status).To(Equal(api.RepositoryHealthy))
				Expect(serviceGetResponse.Service.Versions).To(HaveLen(1))
				Expect(serviceGetResponse.Service.Versions[0]).To(Equal(versionResponse.VersionID))
			})

			It("create version using the same name as the previous creating.", func() {
				version := &api.Version{
					Name:        DefaultVersionName,
					Description: "A version just for test",
					ServiceID:   serviceResponse.ServiceID,
				}
				// Create the version.
				versionResponse = &api.VersionCreationResponse{}

				Expect(CreateVersion(AliceUID, version, versionResponse)).NotTo(BeNil())
				Expect(versionResponse.ErrorMessage).NotTo(Equal(""))
			})
		})

		Context("with missing version id", func() {
			It("should return: unfound log", func() {
				err := WatchLog(ws, APICreateVersion, AliceUID,
					serviceResponse.ServiceID, "-1")
				Expect(err).To(Equal(ErrUnfoundLog))
			})
		})
	})

	Context("with wrong service id", func() {
		It("should fail to create the version, because cyclone couldn't find the service by serviceID.", func() {
			version := &api.Version{
				Name:        DefaultVersionName,
				Description: "A version just for test",
				ServiceID:   "-1",
				Operation:   api.PublishOperation,
			}

			// Forcibly remove image.
			dockerManager.RemoveImage(dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName, DefaultVersionName))

			// Create the version.
			versionResponse := &api.VersionCreationResponse{}
			Expect(CreateVersion(AliceUID, version, versionResponse)).NotTo(BeNil())
			Expect(versionResponse.ErrorMessage).NotTo(Equal(""))
		})
	})

	Context("with different user ID", func() {
		var (
			versionResponse *api.VersionCreationResponse
		)

		It("should create version successfully.", func() {
			version := &api.Version{
				Name:        SecondVersionName,
				Description: "A version just for test",
				ServiceID:   serviceResponse.ServiceID,
				Operation:   api.PublishOperation,
			}

			imageName := dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName, SecondVersionName)
			// Forcibly remove image.
			dockerManager.RemoveImage(imageName)
			// Create the version.
			versionResponse = &api.VersionCreationResponse{}

			Expect(CreateVersion(AliceUID, version, versionResponse)).To(BeNil())
			Expect(versionResponse.ErrorMessage).To(Equal(""))
		})

		It("should build a docker image.", func() {
			versionGetResponse := &api.VersionGetResponse{}
			// Wait up to 120 seconds until docker image has been pushed.
			err := wait.Poll(3*time.Second, 120*time.Second, func() (bool, error) {
				err := GetVersion(AliceUID, versionResponse.VersionID, versionGetResponse)
				return versionGetResponse.Version.Status != api.VersionPending && versionGetResponse.Version.Status != api.VersionRunning, err
			})
			Expect(err).To(BeNil())
			Expect(versionGetResponse.Version.Status).To(Equal(api.VersionHealthy))

			imageName := dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName, SecondVersionName)
			ok, err := dockerManager.IsImagePresent(imageName)
			Expect(ok).To(Equal(false))
			Expect(err).To(BeNil())
			Expect(dockerManager.PullImage(imageName)).To(BeNil())
			ok, err = dockerManager.IsImagePresent(imageName)
			Expect(ok).To(Equal(true))
			Expect(err).To(BeNil())
		})

		It("should not be got by HTTP GET method due to different user ID.", func() {
			versionGetResponse := &api.VersionGetResponse{}
			Expect(GetVersion(BobUID, versionResponse.VersionID, versionGetResponse)).NotTo(BeNil())
		})

		It("should not be able to list version via HTTP GET method due to different user ID.", func() {
			versionListResponse := &api.VersionListResponse{}
			Expect(ListVersions(BobUID, serviceResponse.ServiceID, versionListResponse)).NotTo(BeNil())
		})

		It("should not be able to keep the log due to different user ID.", func() {
			err := WatchLog(ws, APICreateVersion, BobUID, serviceResponse.ServiceID,
				versionResponse.VersionID)
			Expect(err).To(Equal(ErrUnfoundLog))
		})
	})

	Context("should be deleted by HTTP DELETE method. ", func() {
		It("should be able to delete versions and logs via HTTP DELETE method.", func() {
			versionListResponse := &api.VersionListResponse{}
			Expect(ListVersions(AliceUID, serviceResponse.ServiceID, versionListResponse)).To(BeNil())
			Expect(versionListResponse.Versions).To(HaveLen(2))
			Expect(versionListResponse.Versions[0].Name).To(Equal(SecondVersionName))
			Expect(versionListResponse.Versions[0].Status).To(Equal(api.VersionHealthy))
			Expect(versionListResponse.Versions[1].Name).To(Equal(DefaultVersionName))
			Expect(versionListResponse.Versions[1].Status).To(Equal(api.VersionHealthy))
			// Delete service、versions and logs
			serviceDelResponse := &api.ServiceDelResponse{}
			err := DeleteService(AliceUID, serviceResponse.ServiceID, serviceDelResponse)
			Expect(err).To(BeNil())
			// Check if the service is deleted from DB
			serviceGetResponse := &api.ServiceGetResponse{}
			err = GetService(AliceUID, serviceResponse.ServiceID, serviceGetResponse)
			Expect(err).NotTo(BeNil())
			// Check if the versions are deleted from DB
			versionGetResponse := &api.VersionGetResponse{}
			err = GetVersion(AliceUID, versionListResponse.Versions[0].VersionID, versionGetResponse)
			Expect(err).NotTo(BeNil())
			err = GetVersion(AliceUID, versionListResponse.Versions[1].VersionID, versionGetResponse)
			Expect(err).NotTo(BeNil())
			// Check if the logs are still existed on the disk
			statusCode, err := GetVersionLogs(AliceUID, versionListResponse.Versions[0].VersionID)
			Expect(statusCode).To(Equal(http.StatusNotFound)) //can't go pass checkACLForVersion
			Expect(err).To(BeNil())
			statusCode, err = GetVersionLogs(AliceUID, versionListResponse.Versions[1].VersionID)
			Expect(statusCode).To(Equal(http.StatusNotFound)) //can't go pass checkACLForVersion
			Expect(err).To(BeNil())
		})
	})
	Context("test svn vcs", func() {
		// Create a svn test service.
		var versionResponse *api.VersionCreationResponse
		It("should create version successfully.", func() {
			version := &api.Version{
				Name:        DefaultVersionName + "_svn",
				Description: "A version just for svn test",
				ServiceID:   serviceResponse_svn.ServiceID,
				Operation:   api.PublishOperation,
			}

			imageName := dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName+"_svn", DefaultVersionName+"_svn")
			dockerManager.RemoveImage(imageName)
			versionResponse = &api.VersionCreationResponse{}

			Expect(CreateVersion(AliceUID, version, versionResponse)).To(BeNil())
			Expect(versionResponse.ErrorMessage).To(Equal(""))
		})

		It("should be able to get version via HTTP GET method.", func() {
			versionGetResponse := &api.VersionGetResponse{}
			// Wait up to 120 seconds until docker image has been pushed.
			err := wait.Poll(3*time.Second, 120*time.Second, func() (bool, error) {
				err := GetVersion(AliceUID, versionResponse.VersionID, versionGetResponse)
				return versionGetResponse.Version.Status != api.VersionPending && versionGetResponse.Version.Status != api.VersionRunning, err
			})
			Expect(err).To(BeNil())
			Expect(versionGetResponse.ErrorMessage).To(Equal(""))
			Expect(versionGetResponse.Version).NotTo(BeNil())
			Expect(versionGetResponse.Version.Status).To(Equal(api.VersionHealthy))
		})

		It("should clean up docker images from build process.", func() {
			imageName := dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName+"_svn", DefaultVersionName+"_svn")
			ok, err := dockerManager.IsImagePresent(imageName)
			Expect(ok).To(Equal(false))
			Expect(err).To(BeNil())
		})

		It("should pull newly created docker image successfully.", func() {
			imageName := dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName+"_svn", DefaultVersionName+"_svn")
			Expect(dockerManager.PullImage(imageName)).To(BeNil())
			ok, err := dockerManager.IsImagePresent(imageName)
			Expect(ok).To(Equal(true))
			Expect(err).To(BeNil())
		})

		It("should be able to list version via HTTP GET method.", func() {
			versionListResponse := &api.VersionListResponse{}
			Expect(ListVersions(AliceUID, serviceResponse_svn.ServiceID, versionListResponse)).To(BeNil())
			Expect(versionListResponse.Versions).To(HaveLen(1))
			Expect(versionListResponse.Versions[0].Name).To(Equal(DefaultVersionName + "_svn"))
			Expect(versionListResponse.Versions[0].Status).To(Equal(api.VersionHealthy))
		})
	})
})
