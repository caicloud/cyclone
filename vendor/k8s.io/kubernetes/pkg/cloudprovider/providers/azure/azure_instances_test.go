/*
Copyright 2018 The Kubernetes Authors.

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

package azure

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/apimachinery/pkg/types"
)

// setTestVirtualMachines sets test virtual machine with powerstate.
func setTestVirtualMachines(c *Cloud, vmList map[string]string) {
	virtualMachineClient := c.VirtualMachinesClient.(*fakeAzureVirtualMachinesClient)
	store := map[string]map[string]compute.VirtualMachine{
		"rg": make(map[string]compute.VirtualMachine),
	}

	for nodeName, powerState := range vmList {
		instanceID := fmt.Sprintf("/subscriptions/subscription/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/%s", nodeName)
		vm := compute.VirtualMachine{
			Name:     &nodeName,
			ID:       &instanceID,
			Location: &c.Location,
		}
		if powerState != "" {
			status := []compute.InstanceViewStatus{
				{
					Code: to.StringPtr(powerState),
				},
				{
					Code: to.StringPtr("ProvisioningState/succeeded"),
				},
			}
			vm.VirtualMachineProperties = &compute.VirtualMachineProperties{
				InstanceView: &compute.VirtualMachineInstanceView{
					Statuses: &status,
				},
			}
		}

		store["rg"][nodeName] = vm
	}

	virtualMachineClient.setFakeStore(store)
}

func TestInstanceID(t *testing.T) {
	cloud := getTestCloud()
	cloud.metadata = &InstanceMetadata{}

	testcases := []struct {
		name         string
		vmList       []string
		nodeName     string
		metadataName string
		expected     string
		expectError  bool
	}{
		{
			name:         "InstanceID should get instanceID if node's name are equal to metadataName",
			vmList:       []string{"vm1"},
			nodeName:     "vm1",
			metadataName: "vm1",
			expected:     "/subscriptions/subscription/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1",
		},
		{
			name:         "InstanceID should get instanceID from Azure API if node is not local instance",
			vmList:       []string{"vm2"},
			nodeName:     "vm2",
			metadataName: "vm1",
			expected:     "/subscriptions/subscription/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm2",
		},
		{
			name:        "InstanceID should report error if VM doesn't exist",
			vmList:      []string{"vm1"},
			nodeName:    "vm3",
			expectError: true,
		},
	}

	for _, test := range testcases {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Errorf("Test [%s] unexpected error: %v", test.name, err)
		}

		mux := http.NewServeMux()
		mux.Handle("/instance/compute", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, fmt.Sprintf("{\"name\":\"%s\"}", test.metadataName))
		}))
		go func() {
			http.Serve(listener, mux)
		}()
		defer listener.Close()

		cloud.metadata.baseURL = "http://" + listener.Addr().String() + "/"
		vmListWithPowerState := make(map[string]string)
		for _, vm := range test.vmList {
			vmListWithPowerState[vm] = ""
		}
		setTestVirtualMachines(cloud, vmListWithPowerState)
		instanceID, err := cloud.InstanceID(context.Background(), types.NodeName(test.nodeName))
		if test.expectError {
			if err == nil {
				t.Errorf("Test [%s] unexpected nil err", test.name)
			}
		} else {
			if err != nil {
				t.Errorf("Test [%s] unexpected error: %v", test.name, err)
			}
		}

		if instanceID != test.expected {
			t.Errorf("Test [%s] unexpected instanceID: %s, expected %q", test.name, instanceID, test.expected)
		}
	}
}

func TestInstanceShutdownByProviderID(t *testing.T) {
	testcases := []struct {
		name        string
		vmList      map[string]string
		nodeName    string
		expected    bool
		expectError bool
	}{
		{
			name:     "InstanceShutdownByProviderID should return false if the vm is in PowerState/Running status",
			vmList:   map[string]string{"vm1": "PowerState/Running"},
			nodeName: "vm1",
			expected: false,
		},
		{
			name:     "InstanceShutdownByProviderID should return true if the vm is in PowerState/Deallocated status",
			vmList:   map[string]string{"vm2": "PowerState/Deallocated"},
			nodeName: "vm2",
			expected: true,
		},
		{
			name:     "InstanceShutdownByProviderID should return false if the vm is in PowerState/Deallocating status",
			vmList:   map[string]string{"vm3": "PowerState/Deallocating"},
			nodeName: "vm3",
			expected: false,
		},
		{
			name:     "InstanceShutdownByProviderID should return false if the vm is in PowerState/Starting status",
			vmList:   map[string]string{"vm4": "PowerState/Starting"},
			nodeName: "vm4",
			expected: false,
		},
		{
			name:     "InstanceShutdownByProviderID should return true if the vm is in PowerState/Stopped status",
			vmList:   map[string]string{"vm5": "PowerState/Stopped"},
			nodeName: "vm5",
			expected: true,
		},
		{
			name:     "InstanceShutdownByProviderID should return false if the vm is in PowerState/Stopping status",
			vmList:   map[string]string{"vm6": "PowerState/Stopping"},
			nodeName: "vm6",
			expected: false,
		},
		{
			name:     "InstanceShutdownByProviderID should return false if the vm is in PowerState/Unknown status",
			vmList:   map[string]string{"vm7": "PowerState/Unknown"},
			nodeName: "vm7",
			expected: false,
		},
		{
			name:        "InstanceShutdownByProviderID should report error if VM doesn't exist",
			vmList:      map[string]string{"vm1": "PowerState/running"},
			nodeName:    "vm8",
			expectError: true,
		},
	}

	for _, test := range testcases {
		cloud := getTestCloud()
		setTestVirtualMachines(cloud, test.vmList)
		providerID := "azure://" + cloud.getStandardMachineID("rg", test.nodeName)
		hasShutdown, err := cloud.InstanceShutdownByProviderID(context.Background(), providerID)
		if test.expectError {
			if err == nil {
				t.Errorf("Test [%s] unexpected nil err", test.name)
			}
		} else {
			if err != nil {
				t.Errorf("Test [%s] unexpected error: %v", test.name, err)
			}
		}

		if hasShutdown != test.expected {
			t.Errorf("Test [%s] unexpected hasShutdown: %v, expected %v", test.name, hasShutdown, test.expected)
		}
	}
}
