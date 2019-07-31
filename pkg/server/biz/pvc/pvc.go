package pvc

import (
	"time"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/util/retry"
)

// CreatePVC creates pvc for tenant
func CreatePVC(tenantName, storageClass, size string, namespace string, client *kubernetes.Clientset) error {
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
	//err = usage.LaunchPVCUsageWatcher(client, tenantName, v1alpha1.ExecutionContext{
	//	Namespace: nsname,
	//	PVC:       pvcName,
	//})
	//if err != nil {
	//	log.Warningf("Launch PVC usage watcher for %s/%s error: %v", nsname, pvcName, err)
	//}

	return nil
}

// DeletePVC delete the pvc
func DeletePVC(tenantName, namespace string, client *kubernetes.Clientset) error {
	pvcName := common.TenantPVC(tenantName)
	nsname := common.TenantNamespace(tenantName)
	if namespace != "" {
		nsname = namespace
	}

	//err := usage.DeletePVCUsageWatcher(client, nsname)
	//if err != nil && !errors.IsNotFound(err) {
	//	log.Errorf("delete pvc usage watcher in namespace %s failed %s", nsname, err)
	//	return err
	//}

	err := client.CoreV1().PersistentVolumeClaims(nsname).Delete(pvcName, &meta_v1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Errorf("delete persistent volume claim %s error %v", pvcName, err)
		return err
	}

	return nil
}

// UpdatePVC delete the old pvc and recreate another one, so the data of the pvc will lost.
func UpdatePVC(tenantName, storageClass, size string, namespace string, client *kubernetes.Clientset) error {
	err := DeletePVC(tenantName, namespace, client)
	if err != nil {
		return err
	}

	backoff := wait.Backoff{
		Steps:    18,
		Duration: 10 * time.Second,
		Factor:   1.0,
		Jitter:   0.1,
	}

	err = retry.OnError(backoff, func() error {
		return CreatePVC(tenantName, storageClass, size, namespace, client)
	})
	if err != nil {
		return err
	}

	return nil
}
