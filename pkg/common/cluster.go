package common

import (
	"fmt"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MainKubeClient clientset.Interface

// InControlClusterVPC judge the cluster this wfr will be executed weather in the same vpc of ControlCluster
func InControlClusterVPC(name string) bool {
	if name == "" || name == ControlClusterName {
		return true
	}
	// not control cluster, judge weather this cluster's eip equal to vip, if so , this cluster is in
	// the same vpc of controller cluster
	cluster, err := MainKubeClient.CaicloudV1beta1().Clusters().Get(name, v1.GetOptions{})
	if err != nil {
		fmt.Println("Get cluster err is:", err)
		return false
	}
	return cluster.Spec.MastersEIP == cluster.Spec.MastersVIP

	return false
}
