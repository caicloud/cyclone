package common

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"github.com/caicloud/storage-admin/pkg/cluster"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

type Watcher struct {
	clusterAdmAddr     string
	clusterAdmWatchSec int

	ccg kubernetes.ClusterClientsetGetter

	processUpdateList func(map[string]kubernetes.Interface, []string)
	processAddList    func(map[string]kubernetes.Interface, []string)
	processDelList    func(map[string]kubernetes.Interface, []string)
}

func NewWatcher(clusterAdmAddr string, clusterAdmWatchSec int) (*Watcher, error) {
	cm, e := cluster.GetKubeClientsFromClusterAdmin(clusterAdmAddr)
	if e != nil {
		return nil, fmt.Errorf("get kube clientset from cluster admin failed, %v", e)
	}
	ccg := kubernetes.NewClusterClientsetGetter(cm)

	if clusterAdmWatchSec < 1 {
		return nil, fmt.Errorf("bad cluster admin watch second %d", clusterAdmWatchSec)
	}

	w := &Watcher{
		ccg:                ccg,
		clusterAdmAddr:     clusterAdmAddr,
		clusterAdmWatchSec: clusterAdmWatchSec,
	}
	w.SetProcessors(w.ProcessUpdateListSimple, w.ProcessAddListSimple, w.ProcessDelListSimple)
	return w, nil
}

func (w *Watcher) SetProcessors(update, add, del func(map[string]kubernetes.Interface, []string)) {
	w.processUpdateList = update
	w.processAddList = add
	w.processDelList = del
}

func (w *Watcher) Start(stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			glog.Infof("server stopped, watchClusterAdmin exit")
			break
		case <-time.After(time.Duration(w.clusterAdmWatchSec) * time.Second):
		}
		cm, e := cluster.GetKubeClientsFromClusterAdmin(w.clusterAdmAddr)
		if e != nil {
			glog.Errorf("watchClusterAdmin GetKubeClientsFromClusterAdmin failed, %v", e)
			continue
		}
		w.updateClusterInfo(cm)
	}
}

func (w *Watcher) updateClusterInfo(cm map[string]kubernetes.Interface) {
	snap := GetClusterClientsetGetterSnapshot(w.ccg)
	addList, delList, updateList := GetClusterClientsetMapDiff(snap, cm)

	// update
	if w.processUpdateList != nil {
		w.processUpdateList(cm, updateList)
	}
	// add
	if w.processAddList != nil {
		w.processAddList(cm, addList)
	}
	// del
	if w.processDelList != nil {
		w.processDelList(cm, delList)
	}
}

func (w *Watcher) ClusterClientsetGetter() kubernetes.ClusterClientsetGetter {
	return w.ccg
}

func (w *Watcher) SetWatchSec(sec int) {
	w.clusterAdmWatchSec = sec
}

func (w *Watcher) SetWatchAddr(addr string) {
	w.clusterAdmAddr = addr
}

func (w *Watcher) ProcessUpdateListSimple(cm map[string]kubernetes.Interface, updateList []string) {
	for _, cluster := range updateList {
		nc := cm[cluster]
		pc := w.ccg.Get(cluster)
		if nc == nil {
			glog.Errorf("update cluster %s clientset failed, new clientset is nil", cluster)
			continue
		}
		if pc == nil {
			glog.Errorf("update cluster %s clientset failed, prev clientset not exist", cluster)
			continue
		}
		w.ccg.Put(cluster, nc)
		glog.Infof("update cluster %s clientset done", cluster)
	}
}

func (w *Watcher) ProcessAddListSimple(cm map[string]kubernetes.Interface, addList []string) {
	for _, cluster := range addList {
		nc := cm[cluster]
		pc := w.ccg.Get(cluster)
		if nc == nil {
			glog.Errorf("add cluster %s clientset failed, new clientset is nil", cluster)
			continue
		}
		if pc != nil {
			glog.Errorf("add cluster %s clientset failed, prev clientset exist", cluster)
			continue
		}

		// add clientset in ccg
		w.ccg.Put(cluster, nc)
		glog.Infof("add cluster %s clientset done", cluster)
	}
}

func (w *Watcher) ProcessDelListSimple(cm map[string]kubernetes.Interface, delList []string) {
	for _, cluster := range delList {
		nc := cm[cluster]
		pc := w.ccg.Get(cluster)
		if nc != nil {
			glog.Errorf("del cluster %s clientset failed, new clientset exist", cluster)
			continue
		}
		if pc == nil {
			glog.Warningf("del cluster %s clientset failed, prev clientset is nil", cluster)
		}

		w.ccg.Del(cluster)
		glog.Infof("del cluster %s clientset done", cluster)
	}
}

func GetClusterClientsetGetterSnapshot(ccg kubernetes.ClusterClientsetGetter) map[string]kubernetes.Interface {
	snap := make(map[string]kubernetes.Interface)
	if ccg != nil {
		ccg.Range(func(cluster string, kc kubernetes.Interface) error {
			snap[cluster] = kc
			return nil
		})
	}
	return snap
}

func GetClusterClientsetMapDiff(prev, next map[string]kubernetes.Interface) (addList, delList, updateList []string) {
	for cluster, pc := range prev {
		nc := next[cluster]
		if nc == nil {
			// some cluster not exist
			delList = append(delList, cluster)
		} else if !kubernetes.IsConfigSame(pc.RestConfig(), nc.RestConfig()) {
			// some cluster updated
			updateList = append(updateList, cluster)
		}
	}
	for cluster := range next {
		if prev[cluster] == nil {
			// some cluster created
			addList = append(addList, cluster)
		}
	}
	return
}
