package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/caicloud/nirvana/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/integration/cluster"
	svrcommon "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	wfcommon "github.com/caicloud/cyclone/pkg/workflow/common"
)

const (
	// GCImageName is name of the GC image in config
	GCImageName = "gc"
	// GCDefaultImage is the default image used for GC pod
	GCDefaultImage = "alpine:3.8"
)

// GetStorageUsage gets storage usage of the tenant
func GetStorageUsage(ctx context.Context, tenant string) (*v1alpha1.StorageUsage, error) {
	ns, err := handler.K8sClient.CoreV1().Namespaces().Get(svrcommon.TenantNamespace(tenant), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get namespace for tenant '%s' error: %v", tenant, err)
	}

	if ns.Annotations == nil || len(ns.Annotations[meta.AnnotationTenantStorageUsage]) == 0 {
		return nil, fmt.Errorf("no annotation %s found in namespace %s", meta.AnnotationTenantStorageUsage, svrcommon.TenantNamespace(tenant))
	}

	data := ns.Annotations[meta.AnnotationTenantStorageUsage]
	usage := &v1alpha1.StorageUsage{}
	err = json.Unmarshal([]byte(data), usage)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %s error: %v", data, err)
	}

	return usage, nil
}

// ReportStorageUsage reports storage usage of a namespace.
func ReportStorageUsage(ctx context.Context, namespace string, request v1alpha1.StorageUsage) error {
	log.Infof("update pvc storage usage, namespace: %s, usage: %s/%s", namespace, request.Used, request.Total)
	b, err := json.Marshal(request)
	if err != nil {
		log.Warningf("Marshal usage error: %v", err)
		return fmt.Errorf("marshal usage error: %v", err)
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ns, err := handler.K8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if err != nil {
			log.Errorf("Get namespace '%s' error: %v", namespace, err)
			return err
		}

		if ns.Annotations == nil {
			ns.Annotations = make(map[string]string)
		}

		ns.Annotations[meta.AnnotationTenantStorageUsage] = string(b)

		_, err = handler.K8sClient.CoreV1().Namespaces().Update(ns)
		if err != nil {
			log.Warningf("Update namespace '%s' error: %v", namespace, err)
		}
		return err
	})
}

// Cleanup cleans up given storage paths.
func Cleanup(ctx context.Context, tenant string, request v1alpha1.StorageCleanup) error {
	integrations, err := cluster.GetSchedulableClusters(handler.K8sClient, tenant)
	if err != nil {
		return err
	}
	if len(integrations) != 1 {
		return fmt.Errorf("expect one schedulable cluster, but %d found", len(integrations))
	}

	// TODO(ChenDe): Before cleanup the specific paths, we need to check whether the storage path
	// is being using by pods. If some pods are using them, we can't clean them up.

	cluster := integrations[0].Spec.Cluster
	if cluster == nil {
		return fmt.Errorf("schedulable cluster info is empty for tenant: %s", tenant)
	}

	gcPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("gc-%s-%s", tenant, rand.String(5)),
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				meta.LabelPodKind:      meta.PodKindGC.String(),
				meta.LabelPodCreatedBy: meta.CycloneCreator,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    wfcommon.GCContainerName,
					Image:   imageOrDefault(GCImageName, GCDefaultImage),
					Command: append([]string{"rm", "-rf"}, containerPaths(request.Paths)...),
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      wfcommon.DefaultPvVolumeName,
							MountPath: wfcommon.GCDataPath,
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: wfcommon.DefaultPvVolumeName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: cluster.PVC,
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}

	client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
	if err != nil {
		return fmt.Errorf("create cluster client error: %v", err)
	}

	_, err = client.CoreV1().Pods(cluster.Namespace).Create(gcPod)
	if err != nil {
		return fmt.Errorf("create gc pod error: %v", err)
	}

	return nil
}

func imageOrDefault(name, defaultImage string) string {
	if config.Config.Images == nil {
		return defaultImage
	}

	image := config.Config.Images[name]
	if len(image) == 0 {
		return defaultImage
	}

	return image
}

func containerPaths(paths []string) []string {
	var results []string
	for _, p := range paths {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		results = append(results, wfcommon.GCDataPath+p)
	}

	return results
}
