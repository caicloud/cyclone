package storageclass

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/storage-admin/pkg/controller/common"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

func (c *Controller) processSync(sc *storagev1.StorageClass) error {
	logPath := fmt.Sprintf("scController.processSync[class:%s]", sc.Name)
	// check should we process it
	if !util.IsObjStorageAdminMarked(sc) {
		glog.Infof("%s not marked", logPath)
		return nil
	}
	// clientset
	kc := c.kubeCurClient()
	if kc == nil {
		return fmt.Errorf("can't get kube client")
	}
	// check parameters before update in gateway, not here
	// check finalizer
	if !util.HasClassFinalizer(sc.Finalizers) {
		sc.Finalizers = util.AddClassFinalizer(sc.Finalizers)
		_, e := kc.StorageV1().StorageClasses().Update(sc)
		if e != nil {
			glog.Errorf("%s checkClassFinalizers failed, %v", logPath, e)
			return e
		}
		glog.Infof("%s checkClassFinalizers updated", logPath)
	} else {
		glog.Infof("%s checkClassFinalizers ok", logPath)
	}
	return nil
}

func (c *Controller) processTerminate(sc *storagev1.StorageClass) error {
	logPath := fmt.Sprintf("scController.processTerminate[class:%s]", sc.Name)
	// check should we process it
	if !util.IsObjStorageAdminMarked(sc) {
		glog.Infof("%s not marked", logPath)
		return nil
	}
	// clientset
	kc := c.kubeCurClient()
	if kc == nil {
		return fmt.Errorf("can't get kube client")
	}

	// delete pvc should we do it? -> we must do this, or the resources won't be released, or we set it can't be deleted
	if _, e := c.deletePvcInAllNamespace(sc.Name); e != nil {
		glog.Errorf("%s deletePvcInAllNamespace failed, %v", logPath, e)
		return e
	}
	glog.Infof("%s deletePvcInAllNamespace done", logPath)

	// quota cleanup
	var (
		tnCount int
		ptCount int
		cqCount int
		qtCount int
		e       error
	)
	for i := 0; i < 3 || ptCount+qtCount+tnCount+cqCount > 0; i++ {
		// update partitions
		if ptCount, e = c.updatePartitions(sc.Name); e != nil {
			glog.Errorf("%s updatePartitions failed, %v", logPath, e)
			return e
		}
		glog.Infof("%s updatePartitions done %d", logPath, ptCount)

		// update namespace quota
		if qtCount, e = c.updateQuotaInAllNamespace(sc.Name); e != nil {
			glog.Errorf("%s updateQuotaInAllNamespace failed, %v", logPath, e)
			return e
		}
		glog.Infof("%s updateQuotaInAllNamespace done %d", logPath, qtCount)

		// update tenants
		if tnCount, e = c.updateTenants(sc.Name); e != nil {
			glog.Errorf("%s updateTenants failed, %v", logPath, e)
			return e
		}
		glog.Infof("%s updateTenants done %d", logPath, tnCount)

		// update ClusterQuotas
		if cqCount, e = c.updateClusterQuotas(sc.Name); e != nil {
			glog.Errorf("%s updateClusterQuotas failed, %v", logPath, e)
			return e
		}
		glog.Infof("%s updateClusterQuotas done %d", logPath, cqCount)

		glog.Infof("%s in loop %d done %d", logPath, i, tnCount+ptCount+cqCount+qtCount)

		// check and wait
		if i > 10 {
			glog.Errorf("%s updateQuotaInAllNamespace failed, too many repeat: %d, leave it alone", logPath, i)
			break
		} else if i > 0 {
			glog.Infof("%s waiting for next loop", logPath)
			<-time.After(time.Second)
		}
	}

	// delete secret
	secretNamespace, secretName, ok := util.GetStorageSecretPath(sc.Parameters)
	if ok {
		e := common.CheckAndDeleteGlusterfsSecret(kc, secretNamespace, secretName, logPath)
		if e != nil && !kubernetes.IsNotFound(e) {
			glog.Errorf("%s Delete Secrets %s/%s failed, %v", logPath, secretNamespace, secretName, e)
			return e
		}
		glog.Infof("%s cleanup secret %s/%s done", logPath, secretNamespace, secretName)
	}

	// update finalizer
	sc.Finalizers = util.RemoveClassFinalizer(sc.Finalizers)
	_, e = kc.StorageV1().StorageClasses().Update(sc)
	if e != nil && !kubernetes.IsNotFound(e) {
		glog.Errorf("%s Remove Finalizers failed, %v", logPath, e)
		return e
	}
	glog.Infof("%s Remove Finalizers done", logPath)
	return nil
}

func (c *Controller) deletePvcInAllNamespace(storageClassName string) (int, error) {
	kc := c.kubeCurClient()
	if kc == nil {
		return 0, fmt.Errorf("can't get kube client")
	}
	logPrefix := fmt.Sprintf("scController.deletePvcInAllNamespace[class:%s]", storageClassName)

	// list namespace
	nsList, e := kc.CoreV1().Namespaces().List(metav1.ListOptions{})
	if e != nil {
		glog.Errorf("%s Namespaces List failed, %v", logPrefix, e)
		return 0, e
	}
	glog.Infof("%s Namespaces List done", logPrefix)

	allCount := 0
	for _, ns := range nsList.Items {
		logPath := fmt.Sprintf("%s[ns:%s]", logPrefix, ns.Name)
		// list pvc in namespace
		pvcList, e := kc.CoreV1().PersistentVolumeClaims(ns.Name).List(metav1.ListOptions{})
		if e != nil {
			glog.Errorf("%s pvc List failed, %v", logPath, e)
			return allCount, e
		}
		glog.Infof("%s pvc List done", logPath)

		pvcCount := 0
		for i := range pvcList.Items {
			pvc := &pvcList.Items[i]
			// check and delete pvc
			pvcClassName := util.GetPVCClass(pvc)
			if pvcClassName != storageClassName {
				continue
			}
			e = kc.CoreV1().PersistentVolumeClaims(ns.Name).Delete(pvc.Name, nil)
			if e != nil && !kubernetes.IsNotFound(e) {
				glog.Errorf("%s pvc Delete %s failed, %v", logPath, pvc.Name, e)
				return allCount + pvcCount, e
			}
			pvcCount++
			glog.Infof("%s pvc Delete %s done", logPath, pvc.Name)
		}
		glog.Infof("%s pvc Delete done %d", logPath, pvcCount)
		allCount += pvcCount
	}
	glog.Infof("%s pvc Delete done namespace %d", logPrefix, len(nsList.Items))
	return allCount, nil
}

func (c *Controller) updateQuotaInAllNamespace(storageClassName string) (int, error) {
	kc := c.kubeCurClient()
	if kc == nil {
		return 0, fmt.Errorf("can't get kube client")
	}
	logPrefix := fmt.Sprintf("scController.updateQuotaInAllNamespace[class:%s]", storageClassName)
	// add quotas
	nsList, e := kc.CoreV1().Namespaces().List(metav1.ListOptions{})
	if e != nil {
		glog.Errorf("%s Namespaces List failed, %v", logPrefix, e)
		return 0, e
	}

	nsCount := 0
	for _, ns := range nsList.Items {
		quotaName := util.NamespaceQuotaName(ns.Name)
		logPath := fmt.Sprintf("%s Quota %s/%s", logPrefix, ns.Name, quotaName)
		quota, e := kc.CoreV1().ResourceQuotas(ns.Name).Get(quotaName, metav1.GetOptions{})
		if e != nil && !kubernetes.IsNotFound(e) {
			glog.Errorf("%s Get failed, %v", logPath, e)
			return nsCount, e
		}

		if kubernetes.IsNotFound(e) { // quota not exist, should not be this, do this for local test
			glog.Errorf("%s not exist", logPath)
			continue
		}

		if quota.Spec.Hard == nil || len(quota.Spec.Hard) == 0 {
			glog.Warningf("%s is empty", logPath)
			continue
		}
		if needUpdate := deleteStorageClassQuota(storageClassName, quota.Spec.Hard); !needUpdate {
			glog.Infof("%s is up to date", logPath)
			continue
		}
		if _, e = kc.CoreV1().ResourceQuotas(ns.Name).Update(quota); e != nil {
			glog.Errorf("%s Update failed, %v", logPath, e)
			return nsCount, e
		}
		nsCount++
		glog.Infof("%s Update done", logPath)
	}
	glog.Infof("%s Quotas Update done namespace %d", logPrefix, nsCount)
	return nsCount, nil
}

func deleteStorageClassQuota(storageClassName string, rl corev1.ResourceList) (needUpdate bool) {
	if rl == nil {
		return false
	}
	pvcNumLabel := corev1.ResourceName(util.LabelKeyStorageQuotaNum(storageClassName))
	pvcSizeLabel := corev1.ResourceName(util.LabelKeyStorageQuotaSize(storageClassName))
	_, numSet := rl[pvcNumLabel]
	_, sizeSet := rl[pvcSizeLabel]
	if numSet {
		delete(rl, pvcNumLabel)
	}
	if sizeSet {
		delete(rl, pvcSizeLabel)
	}
	needUpdate = numSet || sizeSet
	return needUpdate
}

func deleteStorageClassQuotas(storageClassName string, resLists ...corev1.ResourceList) (needUpdate bool) {
	for _, resList := range resLists {
		resNeedUpdate := deleteStorageClassQuota(storageClassName, resList)
		needUpdate = needUpdate || resNeedUpdate
	}
	return
}

func deleteStorageClassQuotaInRatio(storageClassName string, rl map[corev1.ResourceName]int64) (needUpdate bool) {
	if rl == nil {
		return false
	}
	pvcNumLabel := corev1.ResourceName(util.LabelKeyStorageQuotaNum(storageClassName))
	pvcSizeLabel := corev1.ResourceName(util.LabelKeyStorageQuotaSize(storageClassName))
	_, numSet := rl[pvcNumLabel]
	_, sizeSet := rl[pvcSizeLabel]
	if numSet {
		delete(rl, pvcNumLabel)
	}
	if sizeSet {
		delete(rl, pvcSizeLabel)
	}
	needUpdate = numSet || sizeSet
	return needUpdate
}

func (c *Controller) updateTenants(storageClassName string) (int, error) {
	kc := c.kubeCurClient()
	if kc == nil {
		return 0, fmt.Errorf("can't get kube client")
	}
	logPrefix := fmt.Sprintf("scController.updateTenants[class:%s]", storageClassName)
	// add quotas
	tnList, e := kc.TenantV1alpha1().Tenants().List(metav1.ListOptions{})
	if e != nil {
		glog.Errorf("%s Tenants List failed, %v", logPrefix, e)
		return 0, e
	}

	tnCount := 0
	for i := range tnList.Items {
		tn := &tnList.Items[i]
		logPath := fmt.Sprintf("%s Tenant %s", logPrefix, tn.Name)

		needUpdate := deleteStorageClassQuotas(storageClassName,
			tn.Spec.Quota, tn.Status.Hard, tn.Status.Used, tn.Status.ActualUsed)
		if !needUpdate {
			glog.Infof("%s is up to date", logPath)
			continue
		}

		if _, e = kc.TenantV1alpha1().Tenants().Update(tn); e != nil {
			glog.Errorf("%s Update failed, %v", logPath, e)
			return tnCount, e
		}
		tnCount++
		glog.Infof("%s Update done", logPath)
	}
	glog.Infof("%s Update done %d", logPrefix, tnCount)
	return tnCount, nil
}

func (c *Controller) updatePartitions(storageClassName string) (int, error) {
	kc := c.kubeCurClient()
	if kc == nil {
		return 0, fmt.Errorf("can't get kube client")
	}
	logPrefix := fmt.Sprintf("scController.updatePartitions[class:%s]", storageClassName)
	// add quotas
	ptList, e := kc.TenantV1alpha1().Partitions().List(metav1.ListOptions{})
	if e != nil {
		glog.Errorf("%s Partitions List failed, %v", logPrefix, e)
		return 0, e
	}

	ptCount := 0
	for i := range ptList.Items {
		pt := &ptList.Items[i]
		logPath := fmt.Sprintf("%s Partition %s", logPrefix, pt.Name)

		needUpdate := deleteStorageClassQuotas(storageClassName, pt.Spec.Quota, pt.Status.Hard, pt.Status.Used)
		if !needUpdate {
			glog.Infof("%s is up to date", logPath)
			continue
		}

		if _, e = kc.TenantV1alpha1().Partitions().Update(pt); e != nil {
			glog.Errorf("%s Update failed, %v", logPath, e)
			return ptCount, e
		}
		ptCount++
		glog.Infof("%s Update done", logPath)
	}
	glog.Infof("%s Update done %d", logPrefix, ptCount)
	return ptCount, nil
}

func (c *Controller) updateClusterQuotas(storageClassName string) (int, error) {
	kc := c.kubeCurClient()
	if kc == nil {
		return 0, fmt.Errorf("can't get kube client")
	}
	logPrefix := fmt.Sprintf("scController.updateClusterQuotas[class:%s]", storageClassName)
	// add quotas
	cqList, e := kc.TenantV1alpha1().ClusterQuotas().List(metav1.ListOptions{})
	if e != nil {
		glog.Errorf("%s ClusterQuotas List failed, %v", logPrefix, e)
		return 0, e
	}

	cqCount := 0
	for i := range cqList.Items {
		cq := &cqList.Items[i]
		logPath := fmt.Sprintf("%s ClusterQuotas %s", logPrefix, cq.Name)

		// status
		needUpdate := deleteStorageClassQuotas(storageClassName, cq.Status.Total,
			cq.Status.Allocated, cq.Status.SystemUsed, cq.Status.Used,
			cq.Status.Capacity, cq.Status.Allocatable, cq.Status.Unavailable)
		// ratio
		resNeedUpdate := deleteStorageClassQuotaInRatio(storageClassName, cq.Spec.Ratio)
		// check
		needUpdate = needUpdate || resNeedUpdate
		if !needUpdate {
			glog.Infof("%s is up to date", logPath)
			continue
		}

		if _, e = kc.TenantV1alpha1().ClusterQuotas().Update(cq); e != nil {
			glog.Errorf("%s Update failed, %v", logPath, e)
			return cqCount, e
		}
		cqCount++
		glog.Infof("%s Update done", logPath)
	}
	glog.Infof("%s Update done %d", logPrefix, cqCount)
	return cqCount, nil
}
