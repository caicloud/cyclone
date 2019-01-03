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

package nodeipam

import (
	"net"
	"os"
	"os/exec"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	fakecloud "k8s.io/kubernetes/pkg/cloudprovider/providers/fake"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/nodeipam/ipam"
	"k8s.io/kubernetes/pkg/controller/testutil"
)

func newTestNodeIpamController(clusterCIDR, serviceCIDR *net.IPNet, nodeCIDRMaskSize int, allocatorType ipam.CIDRAllocatorType) (*Controller, error) {
	clientSet := fake.NewSimpleClientset()
	fakeNodeHandler := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "node0"}},
		},
		Clientset: fake.NewSimpleClientset(),
	}
	fakeClient := &fake.Clientset{}
	fakeInformerFactory := informers.NewSharedInformerFactory(fakeClient, controller.NoResyncPeriodFunc())
	fakeNodeInformer := fakeInformerFactory.Core().V1().Nodes()

	for _, node := range fakeNodeHandler.Existing {
		fakeNodeInformer.Informer().GetStore().Add(node)
	}

	fakeCloud := &fakecloud.FakeCloud{}
	return NewNodeIpamController(
		fakeNodeInformer, fakeCloud, clientSet,
		clusterCIDR, serviceCIDR, nodeCIDRMaskSize, allocatorType,
	)
}

// TestNewNodeIpamControllerWithCIDRMasks tests if the controller can be
// created with combinations of network CIDRs and masks.
func TestNewNodeIpamControllerWithCIDRMasks(t *testing.T) {
	for _, tc := range []struct {
		desc          string
		clusterCIDR   string
		serviceCIDR   string
		maskSize      int
		allocatorType ipam.CIDRAllocatorType
		wantFatal     bool
	}{
		{"valid_range_allocator", "10.0.0.0/21", "10.1.0.0/21", 24, ipam.RangeAllocatorType, false},
		{"valid_cloud_allocator", "10.0.0.0/21", "10.1.0.0/21", 24, ipam.CloudAllocatorType, false},
		{"invalid_cluster_CIDR", "invalid", "10.1.0.0/21", 24, ipam.CloudAllocatorType, true},
		{"valid_CIDR_smaller_than_mask_cloud_allocator", "10.0.0.0/26", "10.1.0.0/21", 24, ipam.CloudAllocatorType, false},
		{"invalid_CIDR_smaller_than_mask_other_allocators", "10.0.0.0/26", "10.1.0.0/21", 24, ipam.IPAMFromCloudAllocatorType, true},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, clusterCIDRIpNet, _ := net.ParseCIDR(tc.clusterCIDR)
			_, serviceCIDRIpNet, _ := net.ParseCIDR(tc.serviceCIDR)
			if os.Getenv("EXIT_ON_FATAL") == "1" {
				// This is the subprocess which runs the actual code.
				newTestNodeIpamController(clusterCIDRIpNet, serviceCIDRIpNet, tc.maskSize, tc.allocatorType)
				return
			}
			// This is the host process that monitors the exit code of the subprocess.
			cmd := exec.Command(os.Args[0], "-test.run=TestNewNodeIpamControllerWithCIDRMasks/"+tc.desc)
			cmd.Env = append(os.Environ(), "EXIT_ON_FATAL=1")
			err := cmd.Run()
			var gotFatal bool
			if err != nil {
				exitErr, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatalf("Failed to run subprocess: %v", err)
				}
				gotFatal = !exitErr.Success()
			}
			if gotFatal != tc.wantFatal {
				t.Errorf("newTestNodeIpamController(%v, %v, %v, %v) : gotFatal = %t ; wantFatal = %t", clusterCIDRIpNet, serviceCIDRIpNet, tc.maskSize, tc.allocatorType, gotFatal, tc.wantFatal)
			}
		})
	}
}
