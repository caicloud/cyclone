/*
Copyright 2016 The Kubernetes Authors.

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

package certificates

import (
	"fmt"
	"reflect"
	"strings"

	certificates "k8s.io/kubernetes/pkg/apis/certificates/v1alpha1"
	clientcertificates "k8s.io/kubernetes/pkg/client/clientset_generated/clientset/typed/certificates/v1alpha1"
	certutil "k8s.io/kubernetes/pkg/util/cert"
	utilruntime "k8s.io/kubernetes/pkg/util/runtime"
)

// groupApprover implements AutoApprover for signing Kubelet certificates.
type groupApprover struct {
	client                        clientcertificates.CertificateSigningRequestInterface
	approveAllKubeletCSRsForGroup string
}

// NewGroupApprover creates an approver that accepts any CSR requests where the subject group contains approveAllKubeletCSRsForGroup.
func NewGroupApprover(client clientcertificates.CertificateSigningRequestInterface, approveAllKubeletCSRsForGroup string) AutoApprover {
	return &groupApprover{
		client: client,
		approveAllKubeletCSRsForGroup: approveAllKubeletCSRsForGroup,
	}
}

func (cc *groupApprover) AutoApprove(csr *certificates.CertificateSigningRequest) (*certificates.CertificateSigningRequest, error) {
	// short-circuit if we're not auto-approving
	if cc.approveAllKubeletCSRsForGroup == "" {
		return csr, nil
	}
	// short-circuit if we're already approved or denied
	if approved, denied := getCertApprovalCondition(&csr.Status); approved || denied {
		return csr, nil
	}

	isKubeletBootstrapGroup := false
	for _, g := range csr.Spec.Groups {
		if g == cc.approveAllKubeletCSRsForGroup {
			isKubeletBootstrapGroup = true
			break
		}
	}
	if !isKubeletBootstrapGroup {
		return csr, nil
	}

	x509cr, err := certutil.ParseCSRV1alpha1(csr)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to parse csr %q: %v", csr.Name, err))
		return csr, nil
	}
	if !reflect.DeepEqual([]string{"system:nodes"}, x509cr.Subject.Organization) {
		return csr, nil
	}
	if !strings.HasPrefix(x509cr.Subject.CommonName, "system:node:") {
		return csr, nil
	}
	if len(x509cr.DNSNames)+len(x509cr.EmailAddresses)+len(x509cr.IPAddresses) != 0 {
		return csr, nil
	}

	csr.Status.Conditions = append(csr.Status.Conditions, certificates.CertificateSigningRequestCondition{
		Type:    certificates.CertificateApproved,
		Reason:  "AutoApproved",
		Message: "Auto approving of all kubelet CSRs is enabled on the controller manager",
	})
	return cc.client.UpdateApproval(csr)
}
