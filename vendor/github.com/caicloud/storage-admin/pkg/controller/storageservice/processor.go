package storageservice

import (
	"fmt"
	"time"

	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	"github.com/golang/glog"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/controller/common"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

func (c *Controller) loop(stopCh chan struct{}) {
	tk := time.NewTicker(time.Duration(c.resyncPeriodSecond) * time.Second)
	defer tk.Stop()
	for {
		var svName string
		select {
		case svName = <-c.eventTrigger:
			glog.Infof("loop awake by event trigger from service %v", svName)
		case <-tk.C:
			glog.Infof("loop awake for time up")
			svName = "<time-up>"
		case <-stopCh:
			glog.Infof("loop break for controller stopped")
			return
		}
		c.processSync(svName)
	}
}

func (c *Controller) processSync(serviceName string) error {
	logPrefix := fmt.Sprintf("processSync %s", serviceName)
	cc := c.kubeClient() // control client
	if cc == nil {       // should not be this, just in case
		e := fmt.Errorf("get control cluster client failed")
		glog.Errorf("%s failed, %v", logPrefix, e)
		return e
	}

	c.cleanDeletedService(serviceName, cc)

	processFunc := func(cluster string, kc kubernetes.Interface) error {
		// function will not return any error, for error will stop the big Range
		logPathBase := fmt.Sprintf("%s-[cluster:%s]", logPrefix, cluster)
		scs, e := kc.StorageV1().StorageClasses().List(metav1.ListOptions{})
		if e != nil {
			glog.Errorf("%s List failed, %v", logPathBase, e)
			return nil
		}
		for i := range scs.Items {
			sc := &scs.Items[i]
			logPath := fmt.Sprintf("%s[class:%s]", logPathBase, sc.Name)
			if !util.IsObjStorageAdminMarked(sc) {
				glog.Infof("%s not created by storage admin", logPath)
				continue
			}

			svName := util.GetClassService(sc)
			// should delete it if service not marked? -> has been check in front of this function
			ss, e := getServiceWithCache(svName, c.GetIndexer(), cc)
			if e != nil && !kubernetes.IsNotFound(e) {
				glog.Errorf("%s getService \"%s\" failed, %v", logPath, svName, e)
				continue
			}
			if ss == nil {
				glog.Errorf("%s getService \"%s\" not exist, class will be deleted", logPath, svName)
				e = kc.StorageV1().StorageClasses().Delete(sc.Name, nil)
				if e != nil && !kubernetes.IsNotFound(e) {
					glog.Errorf("%s Delete class for service not found failed, %v", logPath, e)
					continue
				}
				glog.Infof("%s Delete class for service not found done", logPath)
				continue
			}

			// sync parameters
			tp, e := c.tpCache.Get(ss.TypeName)
			if tp == nil {
				glog.Errorf("%s getType failed, %v", logPath, e)
				continue
			}

			needUpdate := syncStorageClassWithTypeAndService(sc, ss, tp)
			if needUpdate {
				glog.Warningf("%s parameters not match", logPath)
			}
		}
		glog.Infof("%s done", logPathBase)
		return nil
	}

	processFunc(constants.ControlClusterName, cc) // for control cluster
	c.ccg.Range(processFunc)                      // for user cluster
	return nil
}

func getServiceWithCache(name string, indexer cache.Indexer, kc kubernetes.Interface) (*resv1b1.StorageService, error) {
	if indexer != nil {
		if obj, exist, e := indexer.GetByKey(name); exist && obj != nil && e == nil {
			if sv, _ := obj.(*resv1b1.StorageService); sv != nil && sv.Name == name {
				return sv, nil
			}
		}
	}
	if kc == nil {
		return nil, fmt.Errorf("service getters nil")
	}
	sv, e := kc.ResourceV1beta1().StorageServices().Get(name, metav1.GetOptions{})
	if e != nil {
		return nil, e
	}
	return sv, nil
}

func syncStorageClassWithTypeAndService(sc *storagev1.StorageClass,
	ss *resv1b1.StorageService, tp *resv1b1.StorageType) (needUpdate bool) {
	newClassParameters := util.SyncStorageClassWithTypeAndService(sc.Parameters, ss.Parameters,
		tp.RequiredParameters, tp.OptionalParameters)
	if newClassParameters == nil {
		return false
	}
	sc.Parameters = newClassParameters
	return true
}

func (c *Controller) cleanDeletedService(serviceName string, cc kubernetes.Interface) error {
	ss, _ := getServiceWithCache(serviceName, c.GetIndexer(), cc)
	if ss == nil || ss.DeletionTimestamp == nil {
		return nil
	}
	logPrefix := fmt.Sprintf("cleanDeletedService %s", serviceName)

	// secret
	secretNamespace, secretName, ok := util.GetStorageSecretPath(ss.Parameters)
	if ok {
		e := common.CheckAndDeleteGlusterfsSecret(cc, secretNamespace, secretName, logPrefix)
		if e != nil && !kubernetes.IsNotFound(e) {
			glog.Errorf("%s cleanup secret %s/%s failed, %v", logPrefix, secretNamespace, secretName, e)
			return e
		}
		glog.Infof("%s cleanup secret %s/%s done", logPrefix, secretNamespace, secretName)
	}

	// finalizer
	ss.Finalizers = util.RemoveServiceFinalizer(ss.Finalizers)
	_, e := cc.ResourceV1beta1().StorageServices().Update(ss)
	if e != nil && !kubernetes.IsNotFound(e) {
		glog.Errorf("%s update finalizers failed, %v", logPrefix, e)
		return e
	}
	glog.Infof("%s update finalizers done", logPrefix)
	return nil
}
