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

package nodelease

import (
	"time"

	coordv1beta1 "k8s.io/api/coordination/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	coordclientset "k8s.io/client-go/kubernetes/typed/coordination/v1beta1"
	"k8s.io/utils/pointer"

	"github.com/golang/glog"
)

const (
	// renewInterval is the interval at which the lease is renewed
	// TODO(mtaufen): 10s was the decision in the KEP, to keep the behavior as close to the
	// current default behavior as possible. In the future, we should determine a reasonable
	// fraction of the lease duration at which to renew, and use that instead.
	renewInterval = 10 * time.Second
	// maxUpdateRetries is the number of immediate, successive retries the Kubelet will attempt
	// when renewing the lease before it waits for the renewal interval before trying again,
	// similar to what we do for node status retries
	maxUpdateRetries = 5
	// maxBackoff is the maximum sleep time during backoff (e.g. in backoffEnsureLease)
	maxBackoff = 7 * time.Second
)

// Controller manages creating and renewing the lease for this Kubelet
type Controller interface {
	Run(stopCh <-chan struct{})
}

type controller struct {
	client                     coordclientset.LeaseInterface
	holderIdentity             string
	leaseDurationSeconds       int32
	renewInterval              time.Duration
	clock                      clock.Clock
	onRepeatedHeartbeatFailure func()
}

// NewController constructs and returns a controller
func NewController(clock clock.Clock, client clientset.Interface, holderIdentity string, leaseDurationSeconds int32, onRepeatedHeartbeatFailure func()) Controller {
	var leaseClient coordclientset.LeaseInterface
	if client != nil {
		leaseClient = client.CoordinationV1beta1().Leases(corev1.NamespaceNodeLease)
	}
	return &controller{
		client:               leaseClient,
		holderIdentity:       holderIdentity,
		leaseDurationSeconds: leaseDurationSeconds,
		renewInterval:        renewInterval,
		clock:                clock,
		onRepeatedHeartbeatFailure: onRepeatedHeartbeatFailure,
	}
}

// Run runs the controller
func (c *controller) Run(stopCh <-chan struct{}) {
	if c.client == nil {
		glog.Infof("node lease controller has nil client, will not claim or renew leases")
		return
	}
	wait.Until(c.sync, c.renewInterval, stopCh)
}

func (c *controller) sync() {
	lease, created := c.backoffEnsureLease()
	// we don't need to update the lease if we just created it
	if !created {
		c.retryUpdateLease(lease)
	}
}

// backoffEnsureLease attempts to create the lease if it does not exist,
// and uses exponentially increasing waits to prevent overloading the API server
// with retries. Returns the lease, and true if this call created the lease,
// false otherwise.
func (c *controller) backoffEnsureLease() (*coordv1beta1.Lease, bool) {
	var (
		lease   *coordv1beta1.Lease
		created bool
		err     error
	)
	sleep := 100 * time.Millisecond
	for {
		lease, created, err = c.ensureLease()
		if err == nil {
			break
		}
		sleep = minDuration(2*sleep, maxBackoff)
		glog.Errorf("failed to ensure node lease exists, will retry in %v, error: %v", sleep, err)
		// backoff wait
		c.clock.Sleep(sleep)
	}
	return lease, created
}

// ensureLease creates the lease if it does not exist. Returns the lease and
// a bool (true if this call created the lease), or any error that occurs.
func (c *controller) ensureLease() (*coordv1beta1.Lease, bool, error) {
	lease, err := c.client.Get(c.holderIdentity, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// lease does not exist, create it
		lease, err := c.client.Create(c.newLease(nil))
		if err != nil {
			return nil, false, err
		}
		return lease, true, nil
	} else if err != nil {
		// unexpected error getting lease
		return nil, false, err
	}
	// lease already existed
	return lease, false, nil
}

// retryUpdateLease attempts to update the lease for maxUpdateRetries,
// call this once you're sure the lease has been created
func (c *controller) retryUpdateLease(base *coordv1beta1.Lease) {
	for i := 0; i < maxUpdateRetries; i++ {
		_, err := c.client.Update(c.newLease(base))
		if err == nil {
			return
		}
		glog.Errorf("failed to update node lease, error: %v", err)
		if i > 0 && c.onRepeatedHeartbeatFailure != nil {
			c.onRepeatedHeartbeatFailure()
		}
	}
	glog.Errorf("failed %d attempts to update node lease, will retry after %v", maxUpdateRetries, c.renewInterval)
}

// newLease constructs a new lease if base is nil, or returns a copy of base
// with desired state asserted on the copy.
func (c *controller) newLease(base *coordv1beta1.Lease) *coordv1beta1.Lease {
	var lease *coordv1beta1.Lease
	if base == nil {
		lease = &coordv1beta1.Lease{}
	} else {
		lease = base.DeepCopy()
	}
	// Use the bare minimum set of fields; other fields exist for debugging/legacy,
	// but we don't need to make node heartbeats more complicated by using them.
	lease.Name = c.holderIdentity
	lease.Spec.HolderIdentity = pointer.StringPtr(c.holderIdentity)
	lease.Spec.LeaseDurationSeconds = pointer.Int32Ptr(c.leaseDurationSeconds)
	lease.Spec.RenewTime = &metav1.MicroTime{Time: c.clock.Now()}
	return lease
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
