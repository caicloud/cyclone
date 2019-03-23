package statistic

import (
	"fmt"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Usage represents usage of some resources, for example storage
type Usage struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
}

// PVCUsage represents PVC usages in a tenant
type PVCUsage struct {
	// Overall is overall usage of the PVC storage
	Overall Usage `json:"overall"`
	// Projects are storage used by each project
	Projects map[string]int64 `json:"projects"`
}

// PVCWatcherName is name of the PVC watcher deployment and pod
const PVCWatcherName = "pvc-watchdog"

// LaunchPVCUsageWatcher launches a pod in a given namespace to report PVC usage regularly.
func LaunchPVCUsageWatcher(client clientset.Interface, context v1alpha1.ExecutionContext) error {
	if len(context.PVC) == 0 {
		return fmt.Errorf("no pvc in execution namespace %s", context.Namespace)
	}

	_, err := client.ExtensionsV1beta1().Deployments(context.Namespace).Create(&v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PVCWatcherName,
			Namespace: context.Namespace,
		},
		Spec: v1beta1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      PVCWatcherName,
					Namespace: context.Namespace,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "c0",
							Image: "insight.caicloudprivatetest.com/release/busybox:1.30.0",
							Command: []string{
								"/bin/sh",
								"-c",
								"",
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "cv",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: context.PVC,
									ReadOnly:  true,
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	})
	return err
}
