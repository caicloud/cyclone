package statistic

import (
	"fmt"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// ReportURLEnvName ...
	ReportURLEnvName = "REPORT_URL"
	// HeartbeatIntervalEnvName ...
	HeartbeatIntervalEnvName = "HEARTBEAT_INTERVAL"
	// NamespaceEnvName ...
	NamespaceEnvName = "NAMESPACE"
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
func LaunchPVCUsageWatcher(client *kubernetes.Clientset, context v1alpha1.ExecutionContext) error {
	if len(context.PVC) == 0 {
		return fmt.Errorf("no pvc in execution namespace %s", context.Namespace)
	}

	watcherConfig := config.Config.StorageUsageWatcher
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
							Image: watcherConfig.Image,
							Env: []corev1.EnvVar{
								{
									Name:  ReportURLEnvName,
									Value: watcherConfig.ReportURL,
								},
								{
									Name:  HeartbeatIntervalEnvName,
									Value: watcherConfig.IntervalSeconds,
								},
								{
									Name:  NamespaceEnvName,
									Value: context.Namespace,
								},
							},
							Command: []string{
								"/bin/sh",
								"-c",
								`while [ true ]; do wget -q --header="Content-Type:application/json" --header="X-Namespace:$NAMESPACE" --post-data="{data: $(stat -tf /data | grep "/data" | awk '{ print $5":"$6":"$7 }')}" $REPORT_URL; sleep $HEARTBEAT_INTERVAL; done`,
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("32Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cv",
									ReadOnly:  true,
									MountPath: "/data",
									SubPath:   "caches",
								},
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
