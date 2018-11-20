package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/golang/glog" // -log_dir=./logs -stderrthreshold=info

	"github.com/caicloud/storage-admin/pkg/admin"
	"github.com/caicloud/storage-admin/pkg/config"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

var (
	clusterAdminHost string
	kubeHost         string
	kubeConfig       string

	controlClusterName string

	kc kubernetes.Interface

	listenPort int

	sigCh  = make(chan os.Signal, 32)
	stopCh = make(chan struct{})
)

func flagInit() {
	flag.StringVar(&clusterAdminHost, config.EnvClusterAdminHost, "", "cluster admin host")
	flag.StringVar(&kubeHost, config.EnvKubeHost, "", "kube config file path")
	flag.StringVar(&kubeConfig, config.EnvKubeConfig, "", "kube host address")
	flag.IntVar(&listenPort, config.EnvListenPort, constants.DefaultListenPort, "http listen port")
	flag.StringVar(&controlClusterName, config.EnvCtrlClusterName, constants.ControlClusterName, "control cluster name, main for test")
	flag.Set(constants.GLogLogLevelFlagName, config.LogLevel)
	flag.Set(constants.GLogLogVerbosityFlagName, config.LogVerbosity)
	flag.Parse()

	if len(config.ClusterAdminHost) > 0 {
		clusterAdminHost = config.ClusterAdminHost
	}
	if len(config.KubeHost) > 0 {
		kubeHost = config.KubeHost
	}
	if len(config.KubeConfig) > 0 {
		kubeConfig = config.KubeConfig
	}
	if config.ListenPort > 0 {
		listenPort = config.ListenPort
	}
	if len(config.CtrlClusterName) > 0 {
		controlClusterName = config.CtrlClusterName
	}

	glog.Infof("arg %s=%v", config.EnvClusterAdminHost, clusterAdminHost)
	glog.Infof("arg %s=%v", config.EnvKubeHost, kubeHost)
	glog.Infof("arg %s=%v", config.EnvKubeConfig, kubeConfig)
	glog.Infof("arg %s=%v", config.EnvCtrlClusterName, controlClusterName)
	glog.Infof("arg %s=%v", config.EnvListenPort, listenPort)
	glog.Infof("arg %s=%v", config.EnvLogLevel, config.LogLevel)
	glog.Infof("arg %s=%v", config.EnvLogVerbosity, config.LogVerbosity)
}

func kubeInit() {
	var e error
	kc, e = kubernetes.NewClientFromFlags(kubeHost, kubeConfig)
	if e != nil {
		glog.Exitf("NewClientFromFlags failed, %v", e)
	}

	if e = kubernetes.InitStorageAdminMetaMainCluster(kc); e != nil {
		glog.Exitf("InitStorageAdminMetaMainCluster failed, %v", e)
	}
	if e = kubernetes.InitDefaultStorageTypes(kc); e != nil {
		glog.Exitf("InitDefaultStorageTypes failed, %v", e)
	}

	glog.Infof("kubeInit cluster done for ctrl cluster only")
}

func main() {
	flagInit()
	kubeInit()
	defer glog.Flush()

	// TODO ccg from cluster admin

	s, e := admin.NewServer(kc, controlClusterName, clusterAdminHost)
	if e != nil {
		glog.Exitf("NewServer failed, %v", e)
	}

	signal.Notify(sigCh, os.Kill)
	go func() {
		<-sigCh
		signal.Stop(sigCh)
		glog.Warningf("receive kill signal, stop program")
		close(stopCh)
	}()

	glog.Infof("Starting storage admin gateway server on port %d", listenPort)
	s.Run(stopCh, listenPort)
}
