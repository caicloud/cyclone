/*
Copyright 2017 The Kubernetes Authors.

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

package uploadconfig

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1alpha3 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1alpha3"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	configutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"
)

func TestUploadConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		errOnCreate    error
		errOnUpdate    error
		updateExisting bool
		errExpected    bool
		verifyResult   bool
	}{
		{
			name:         "basic validation with correct key",
			verifyResult: true,
		},
		{
			name:           "update existing should report no error",
			updateExisting: true,
			verifyResult:   true,
		},
		{
			name:        "unexpected errors for create should be returned",
			errOnCreate: apierrors.NewUnauthorized(""),
			errExpected: true,
		},
		{
			name:           "update existing show report error if unexpected error for update is returned",
			errOnUpdate:    apierrors.NewUnauthorized(""),
			updateExisting: true,
			errExpected:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			initialcfg := &kubeadmapiv1alpha3.InitConfiguration{
				APIEndpoint: kubeadmapiv1alpha3.APIEndpoint{
					AdvertiseAddress: "1.2.3.4",
				},
				ClusterConfiguration: kubeadmapiv1alpha3.ClusterConfiguration{
					KubernetesVersion: "v1.11.10",
				},
				BootstrapTokens: []kubeadmapiv1alpha3.BootstrapToken{
					{
						Token: &kubeadmapiv1alpha3.BootstrapTokenString{
							ID:     "abcdef",
							Secret: "abcdef0123456789",
						},
					},
				},
				NodeRegistration: kubeadmapiv1alpha3.NodeRegistrationOptions{
					Name:      "node-foo",
					CRISocket: "/var/run/custom-cri.sock",
				},
			}
			cfg, err := configutil.ConfigFileAndDefaultsToInternalConfig("", initialcfg)
			if err != nil {
				t2.Fatalf("UploadConfiguration() error = %v", err)
			}

			status := &kubeadmapi.ClusterStatus{
				APIEndpoints: map[string]kubeadmapi.APIEndpoint{
					"node-foo": cfg.APIEndpoint,
				},
			}

			client := clientsetfake.NewSimpleClientset()
			if tt.errOnCreate != nil {
				client.PrependReactor("create", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
					return true, nil, tt.errOnCreate
				})
			}
			// For idempotent test, we check the result of the second call.
			if err := UploadConfiguration(cfg, client); !tt.updateExisting && (err != nil) != tt.errExpected {
				t2.Fatalf("UploadConfiguration() error = %v, wantErr %v", err, tt.errExpected)
			}
			if tt.updateExisting {
				if tt.errOnUpdate != nil {
					client.PrependReactor("update", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
						return true, nil, tt.errOnUpdate
					})
				}
				if err := UploadConfiguration(cfg, client); (err != nil) != tt.errExpected {
					t2.Fatalf("UploadConfiguration() error = %v", err)
				}
			}
			if tt.verifyResult {
				masterCfg, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(kubeadmconstants.InitConfigurationConfigMap, metav1.GetOptions{})
				if err != nil {
					t2.Fatalf("Fail to query ConfigMap error = %v", err)
				}
				configData := masterCfg.Data[kubeadmconstants.ClusterConfigurationConfigMapKey]
				if configData == "" {
					t2.Fatal("Fail to find ClusterConfigurationConfigMapKey key")
				}

				decodedCfg := &kubeadmapi.ClusterConfiguration{}
				if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), []byte(configData), decodedCfg); err != nil {
					t2.Fatalf("unable to decode config from bytes: %v", err)
				}

				if !reflect.DeepEqual(decodedCfg, &cfg.ClusterConfiguration) {
					t2.Errorf("the initial and decoded ClusterConfiguration didn't match")
				}

				statusData := masterCfg.Data[kubeadmconstants.ClusterStatusConfigMapKey]
				if statusData == "" {
					t2.Fatal("failed to find ClusterStatusConfigMapKey key")
				}

				decodedStatus := &kubeadmapi.ClusterStatus{}
				if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), []byte(statusData), decodedStatus); err != nil {
					t2.Fatalf("unable to decode status from bytes: %v", err)
				}

				if !reflect.DeepEqual(decodedStatus, status) {
					t2.Error("the initial and decoded ClusterStatus didn't match")
				}
			}
		})
	}
}

func TestGetClusterStatus(t *testing.T) {
	var tests = []struct {
		name                     string
		clusterStatus            *kubeadmapi.ClusterStatus
		expectedClusterEndpoints int
	}{
		{
			name: "return empty ClusterStatus if cluster kubeadm-config doesn't exist (e.g init)",
			expectedClusterEndpoints: 0,
		},
		{
			name: "return ClusterStatus if cluster kubeadm-config exist (e.g upgrade)",
			clusterStatus: &kubeadmapi.ClusterStatus{
				APIEndpoints: map[string]kubeadmapi.APIEndpoint{
					"dummy": {AdvertiseAddress: "1.2.3.4", BindPort: 1234},
				},
			},
			expectedClusterEndpoints: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := clientsetfake.NewSimpleClientset()

			if tt.clusterStatus != nil {
				createConfigMapWithStatus(tt.clusterStatus, client)
			}

			actual, err := getClusterStatus(client)
			if err != nil {
				t.Error("GetClusterStatus returned unexpected error")
				return
			}
			if tt.expectedClusterEndpoints != len(actual.APIEndpoints) {
				t.Error("actual ClusterStatus doesn't return expected endpoints")
			}
		})
	}
}

// createConfigMapWithStatus create a ConfigMap with ClusterStatus for TestGetClusterStatus
func createConfigMapWithStatus(statusToCreate *kubeadmapi.ClusterStatus, client clientset.Interface) error {
	statusYaml, err := configutil.MarshalKubeadmConfigObject(statusToCreate)
	if err != nil {
		return err
	}

	return apiclient.CreateOrUpdateConfigMap(client, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeadmconstants.InitConfigurationConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			kubeadmconstants.ClusterStatusConfigMapKey: string(statusYaml),
		},
	})
}
