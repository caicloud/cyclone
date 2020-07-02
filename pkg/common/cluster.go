package common

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

// MainKubeClient is clientset of cyclone
var MainKubeClient clientset.Interface

// InControlClusterVPC judge the cluster this wfr will be executed weather in the same vpc of ControlCluster
func InControlClusterVPC(name string) bool {
	if name == "" || name == ControlClusterName {
		return true
	}
	// not control cluster, judge weather this cluster's eip equal to vip, if so , this cluster is in
	// the same vpc of controller cluster
	cluster, err := MainKubeClient.CaicloudV1beta1().Clusters().Get(name, metav1.GetOptions{})
	if err != nil {
		fmt.Println("Get cluster err is:", err)
		return false
	}
	// when EIP = VIP , there is a special case , when cluster running on physical machine and not in
	// cluster VPC, this cluster will use same eip and vip. so judge weather this ip is 10.XXX.XXX.XXX
	// , if so , this is vip
	if cluster.Spec.MastersEIP == cluster.Spec.MastersVIP {
		return strings.HasPrefix(cluster.Spec.MastersEIP, "10.")
	}

	return false
}
