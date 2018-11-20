package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/golang/glog" // -log_dir=./logs -stderrthreshold=info

	"github.com/caicloud/storage-admin/pkg/config"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/controller/common"
	"github.com/caicloud/storage-admin/pkg/controller/storageservice"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

var (
	clusterAdminHost    string
	kubeHost            string
	kubeConfig          string
	watchIntervalSecond int
	maxRetryTimes       int
	resyncPeriodSecond  int

	sigCh  = make(chan os.Signal, 32)
	stopCh = make(chan struct{})
)

func argsInit() {
	flag.StringVar(&clusterAdminHost, config.EnvClusterAdminHost, "", "cluster admin host")
	flag.StringVar(&kubeHost, config.EnvKubeHost, "", "kubernetes host address")
	flag.StringVar(&kubeConfig, config.EnvKubeConfig, "", "kubernetes config path")
	flag.IntVar(&watchIntervalSecond, config.EnvWatchIntervalSecond, 0, "refresh configs from cluster time interval seconds")
	flag.IntVar(&maxRetryTimes, config.EnvControllerMaxRetryTimes, 0, "controller failure max retry time, < 1 means unlimited")
	flag.IntVar(&resyncPeriodSecond, config.EnvControllerResyncPeriodSecond, 0, "controller resync period second, < 1 means no auto resync")
	flag.Set(constants.GLogLogLevelFlagName, config.LogLevel)
	flag.Set(constants.GLogLogVerbosityFlagName, config.LogVerbosity)
	flag.Parse()
	if len(clusterAdminHost) == 0 {
		clusterAdminHost = config.ClusterAdminHost
	}
	if len(kubeHost) == 0 {
		kubeHost = config.KubeHost
	}
	if len(kubeConfig) == 0 {
		kubeConfig = config.KubeConfig
	}
	if watchIntervalSecond == 0 {
		watchIntervalSecond = config.WatchIntervalSecond
	}
	if maxRetryTimes == 0 {
		maxRetryTimes = config.ControllerMaxRetryTimes
	}
	if resyncPeriodSecond == 0 {
		resyncPeriodSecond = config.ControllerResyncPeriodSecond
	}
	glog.Infof("arg %s=%v", config.EnvClusterAdminHost, clusterAdminHost)
	glog.Infof("arg %s=%v", config.EnvKubeHost, kubeHost)
	glog.Infof("arg %s=%v", config.EnvKubeConfig, kubeConfig)
	glog.Infof("arg %s=%v", config.EnvWatchIntervalSecond, watchIntervalSecond)
	glog.Infof("arg %s=%v", config.EnvControllerMaxRetryTimes, maxRetryTimes)
	glog.Infof("arg %s=%v", config.EnvControllerResyncPeriodSecond, resyncPeriodSecond)
	glog.Infof("arg %s=%v", config.EnvLogLevel, config.LogLevel)
	glog.Infof("arg %s=%v", config.EnvLogVerbosity, config.LogVerbosity)
}

func main() {
	argsInit()
	defer glog.Flush()

	// kube client
	kc, e := kubernetes.NewClientFromFlags(kubeHost, kubeConfig)
	if e != nil {
		glog.Exitf("NewClientFromFlags failed, %v", e)
	}

	// watcher for cluster admin
	aw, e := common.NewWatcher(clusterAdminHost, watchIntervalSecond)
	if e != nil {
		glog.Exitf("NewWatcher failed, %v", e)
	}
	aw.SetWatchSec(watchIntervalSecond)

	// make controller
	c, e := storageservice.NewController(kc, aw.ClusterClientsetGetter())
	if e != nil {
		glog.Exitf("NewController failed, %v", e)
	}
	c.SetMaxRetryTimes(maxRetryTimes)
	c.SetResyncPeriodSecond(resyncPeriodSecond)

	// init crd
	if e = c.InitCrd(); e != nil {
		glog.Exitf("InitCrd failed, %v", e)
	}

	stopCh = make(chan struct{})

	signal.Notify(sigCh, os.Kill)
	go func() {
		<-sigCh
		signal.Stop(sigCh)
		glog.Warningf("receive kill signal, stop program")
		close(stopCh)
	}()

	go func() {
		aw.Start(stopCh)
	}()

	go func() {
		c.Run(1, stopCh)
	}()
	<-stopCh
}
