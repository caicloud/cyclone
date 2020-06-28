package pvc

import (
	"fmt"
	"strings"
	"time"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/biz/usage"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	"github.com/caicloud/cyclone/pkg/util/retry"
)

// CreatePVC creates pvc for tenant
func CreatePVC(tenantName, storageClass, size string, namespace string, client *kubernetes.Clientset, isMaster bool) error {
	// parse quantity
	resources := make(map[core_v1.ResourceName]resource.Quantity)
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		log.Errorf("Parse Quantity %s error %v", size, err)
		return err
	}
	resources[core_v1.ResourceStorage] = quantity

	// create pvc
	pvcName := common.TenantPVC(tenantName)
	nsname := common.TenantNamespace(tenantName)
	if namespace != "" {
		nsname = namespace
	}
	volume := &core_v1.PersistentVolumeClaim{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      pvcName,
			Namespace: nsname,
		},
		Spec: core_v1.PersistentVolumeClaimSpec{
			AccessModes: []core_v1.PersistentVolumeAccessMode{core_v1.ReadWriteMany},
			Resources: core_v1.ResourceRequirements{
				Limits:   resources,
				Requests: resources,
			},
		},
	}

	if storageClass != "" {
		volume.Spec.StorageClassName = &storageClass
	}

	_, err = client.CoreV1().PersistentVolumeClaims(nsname).Create(volume)
	if err != nil {
		log.Errorf("Create persistent volume claim %s error %v", pvcName, err)
		return err
	}

	// Launch PVC usage watcher to watch the usage of PVC.
	err = usage.LaunchPVCUsageWatcher(client, tenantName, v1alpha1.ExecutionContext{
		Namespace: nsname,
		PVC:       pvcName,
	}, isMaster)
	if err != nil {
		log.Errorf("Launch PVC usage watcher for %s/%s error: %v", nsname, pvcName, err)
		return err
	}

	return nil
}

// DeletePVC delete the pvc
func DeletePVC(tenantName, namespace string, client *kubernetes.Clientset) error {
	pvcName := common.TenantPVC(tenantName)
	nsname := common.TenantNamespace(tenantName)
	if namespace != "" {
		nsname = namespace
	}

	err := usage.DeletePVCUsageWatcher(client, nsname)
	if err != nil && !errors.IsNotFound(err) {
		log.Errorf("delete pvc usage watcher in namespace %s failed %s", nsname, err)
		return err
	}

	err = client.CoreV1().PersistentVolumeClaims(nsname).Delete(pvcName, &meta_v1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Errorf("delete persistent volume claim %s error %v", pvcName, err)
		return err
	}

	return nil
}

// ConfirmPVCDeleted confirms whether the PVC and PVC watchdog have been deleted.
// Only return nil represents the PVC and PVC watchdog have been deleted.
func ConfirmPVCDeleted(tenantName, namespace string, client *kubernetes.Clientset) error {
	pvcName := common.TenantPVC(tenantName)
	nsname := common.TenantNamespace(tenantName)
	if namespace != "" {
		nsname = namespace
	}

	_, err := usage.GetPVCUsageWatcher(client, nsname)
	if err == nil {
		return fmt.Errorf("PVC watchdog for tenant %s has not been deleted", tenantName)
	}
	if !errors.IsNotFound(err) {
		return err
	}

	_, err = client.CoreV1().PersistentVolumeClaims(nsname).Get(pvcName, meta_v1.GetOptions{})
	if err == nil {
		return fmt.Errorf("PVC for tenant %s has not been deleted", tenantName)
	}
	if !errors.IsNotFound(err) {
		return err
	}

	return nil
}

// UpdatePVC delete the old pvc and recreate another one, so the data of the pvc will lost.
func UpdatePVC(tenantName, storageClass, size string, namespace string, client *kubernetes.Clientset, isMaster bool) error {
	// Can not update pvc when there are workflows running.
	pods, err := client.CoreV1().Pods(namespace).List(meta_v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s!=%s", usage.PVCWatcherLabelName, usage.PVCWatcherLabelValue),
	})
	if err != nil {
		log.Warning("list pvc watcher pods error: ", err)
		return err
	}

	var wfrsName = make(map[string]struct{})
	for _, pod := range pods.Items {
		if pod.Labels != nil && pod.Labels[meta.LabelWorkflowRunName] != "" {
			wfrsName[pod.Labels[meta.LabelWorkflowRunName]] = struct{}{}
		}
	}

	if len(wfrsName) > 0 {
		var wfrsNameString string
		for name := range wfrsName {
			wfrsNameString = fmt.Sprintf("%s,%s", wfrsNameString, name)
		}
		return cerr.ErrorExistRunningWorkflowRuns.Error(strings.Trim(wfrsNameString, ","))
	}

	// Recreate PVC
	err = DeletePVC(tenantName, namespace, client)
	if err != nil {
		return err
	}

	backoff := wait.Backoff{
		Steps:    12,
		Duration: 5 * time.Second,
		Factor:   1.0,
		Jitter:   0.1,
	}

	// Wait for the PVC and PVC watchdog deleting completed
	err = retry.OnError(backoff, func() error {
		return ConfirmPVCDeleted(tenantName, namespace, client)
	})
	if err != nil {
		return err
	}

	// Create a new PVC
	err = retry.OnError(backoff, func() error {
		return CreatePVC(tenantName, storageClass, size, namespace, client, isMaster)
	})
	if err != nil {
		return err
	}

	return nil
}
