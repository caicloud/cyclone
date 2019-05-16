package cluster

import (
	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	rbac_v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/workflow/common"
)

// EnsureServiceAccount ensures default service account for stage pod created. Coordinator container in stage pod
// needs certain permission to pods and pods/log.
func EnsureServiceAccount(clusterClient *kubernetes.Clientset, namespace string) error {
	if _, err := clusterClient.CoreV1().ServiceAccounts(namespace).Create(&core_v1.ServiceAccount{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      common.DefaultServiceAccountName,
			Namespace: namespace,
		},
	}); err != nil {
		if errors.IsAlreadyExists(err) {
			log.Infof("Service account %s already exist, skip it", common.DefaultServiceAccountName)
		} else {
			log.Errorf("Create service account %s error: %v", common.DefaultServiceAccountName, err)
			return err
		}
	}

	// Create PodSecurityPolicy for the service account. Here we create the least restricted policy, it equivalents to
	// not using the pod security policy admission controller.
	var trueValue = true
	if _, err := clusterClient.PolicyV1beta1().PodSecurityPolicies().Create(&v1beta1.PodSecurityPolicy{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      common.DefaultServiceAccountName,
			Namespace: namespace,
		},
		Spec: v1beta1.PodSecurityPolicySpec{
			Privileged:               true,
			AllowPrivilegeEscalation: &trueValue,
			AllowedCapabilities:      []core_v1.Capability{"*"},
			HostNetwork:              true,
			HostPID:                  true,
			HostIPC:                  true,
			HostPorts:                []v1beta1.HostPortRange{{Min: 0, Max: 65535}},
			Volumes:                  []v1beta1.FSType{"*"},
			RunAsUser:                v1beta1.RunAsUserStrategyOptions{Rule: v1beta1.RunAsUserStrategyRunAsAny},
			SELinux:                  v1beta1.SELinuxStrategyOptions{Rule: v1beta1.SELinuxStrategyRunAsAny},
			SupplementalGroups:       v1beta1.SupplementalGroupsStrategyOptions{Rule: v1beta1.SupplementalGroupsStrategyRunAsAny},
			FSGroup:                  v1beta1.FSGroupStrategyOptions{Rule: v1beta1.FSGroupStrategyRunAsAny},
		},
	}); err != nil {
		log.Warningf("Create PodSecurityPolicy error: %s, if cluster has pod security policy admission controller enabled, stage pod may fail", err)
	}

	if _, err := clusterClient.RbacV1().ClusterRoles().Create(&rbac_v1.ClusterRole{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: common.DefaultServiceAccountName,
		},
		Rules: []rbac_v1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/log"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups:     []string{"policy"},
				Resources:     []string{"podsecuritypolicies"},
				Verbs:         []string{"use"},
				ResourceNames: []string{common.DefaultServiceAccountName},
			},
		},
	}); err != nil {
		if errors.IsAlreadyExists(err) {
			log.Infof("ClusterRole %s already exist, skip it", common.DefaultServiceAccountName)
		} else {
			log.Errorf("Create ClusterRole %s error: %v", common.DefaultServiceAccountName, err)
			return err
		}
	}

	if _, err := clusterClient.RbacV1().RoleBindings(namespace).Create(&rbac_v1.RoleBinding{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      common.DefaultServiceAccountName,
			Namespace: namespace,
		},
		Subjects: []rbac_v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      common.DefaultServiceAccountName,
				Namespace: namespace,
			},
		},
		RoleRef: rbac_v1.RoleRef{
			Kind:     "ClusterRole",
			Name:     common.DefaultServiceAccountName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}); err != nil {
		if errors.IsAlreadyExists(err) {
			log.Infof("RoleBingding %s already exist, skip it", common.DefaultServiceAccountName)
		} else {
			log.Errorf("Create RoleBinding %s error: %v", common.DefaultServiceAccountName, err)
			return err
		}
	}

	return nil
}
