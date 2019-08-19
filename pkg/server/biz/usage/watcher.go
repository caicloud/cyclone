package usage

import (
	"fmt"

	"github.com/caicloud/nirvana/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
)

const (
	// ReportURLEnvName ...
	ReportURLEnvName = "REPORT_URL"
	// HeartbeatIntervalEnvName ...
	HeartbeatIntervalEnvName = "HEARTBEAT_INTERVAL"
	// NamespaceEnvName ...
	NamespaceEnvName = "NAMESPACE"

	// PVCWatcherLabelName ...
	PVCWatcherLabelName = "pod.cyclone.dev/name"
	// PVCWatcherLabelValue ...
	PVCWatcherLabelValue = "pvc-watcher"
)

var (
	// K8sClient is used to operate k8s resources in control cluster
	controlClusterClient kubernetes.Interface
)

// Init initializes the control cluster client.
func Init(c kubernetes.Interface) {
	controlClusterClient = c
}

// PVCWatcherName is name of the PVC watcher deployment and pod
const PVCWatcherName = "pvc-watchdog"

// WatcherController controls pvc watcher
type WatcherController struct {
	tenant     string
	userClient kubernetes.Interface
	recorder   Recorder
}

// NewWatcherController creates a PVC WatcherController
func NewWatcherController(client kubernetes.Interface, tenant string) *WatcherController {
	return &WatcherController{
		userClient: client,
		tenant:     tenant,
		recorder:   NewNamespaceRecorder(controlClusterClient, common.TenantNamespace(tenant)),
	}
}

// LaunchPVCUsageWatcher launches a pod in a given namespace to report PVC usage regularly.
func (w *WatcherController) LaunchPVCUsageWatcher(context v1alpha1.ExecutionContext) error {
	if len(context.PVC) == 0 {
		return fmt.Errorf("no pvc in execution namespace %s", context.Namespace)
	}

	watcherConfig := config.Config.StorageUsageWatcher
	resourceRequirements := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(getOrDefault(&watcherConfig, corev1.ResourceRequestsCPU, "50m")),
			corev1.ResourceMemory: resource.MustParse(getOrDefault(&watcherConfig, corev1.ResourceRequestsMemory, "32Mi")),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(getOrDefault(&watcherConfig, corev1.ResourceLimitsCPU, "100m")),
			corev1.ResourceMemory: resource.MustParse(getOrDefault(&watcherConfig, corev1.ResourceLimitsMemory, "128Mi")),
		},
	}

	_, err := w.userClient.ExtensionsV1beta1().Deployments(context.Namespace).Create(&v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PVCWatcherName,
			Namespace: context.Namespace,
		},
		Spec: v1beta1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					PVCWatcherLabelName: PVCWatcherLabelValue,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      PVCWatcherName,
					Namespace: context.Namespace,
					Labels: map[string]string{
						PVCWatcherLabelName: PVCWatcherLabelValue,
					},
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
									Value: common.TenantNamespace(w.tenant),
								},
							},
							Resources: resourceRequirements,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cv",
									ReadOnly:  true,
									MountPath: "/pvc-data",
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

	if recordErr := w.recorder.RecordWatcherResource(resourceRequirements); err != nil {
		log.Warningf("record pvc watcher resource requirements error: %v", recordErr)
	}
	return err
}

// getOrDefault gets resource requirement from config, if not set, use default value.
func getOrDefault(watcherConfig *config.StorageUsageWatcher, key corev1.ResourceName, defaultValue string) string {
	v, ok := watcherConfig.ResourceRequirements[key]
	if ok {
		return v
	}

	return defaultValue
}

// DeletePVCUsageWatcher delete the pvc usage watcher deployment
func (w *WatcherController) DeletePVCUsageWatcher(namespace string) error {
	foreground := metav1.DeletePropagationForeground
	err := w.userClient.ExtensionsV1beta1().Deployments(namespace).Delete(PVCWatcherName, &metav1.DeleteOptions{
		PropagationPolicy: &foreground,
	})

	zeroQuantity := resource.MustParse("0")
	resourceRequirements := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    zeroQuantity,
			corev1.ResourceMemory: zeroQuantity,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    zeroQuantity,
			corev1.ResourceMemory: zeroQuantity,
		},
	}
	if recordErr := w.recorder.RecordWatcherResource(resourceRequirements); err != nil {
		log.Warningf("record pvc watcher resource requirements error: %v", recordErr)
	}

	return err
}
