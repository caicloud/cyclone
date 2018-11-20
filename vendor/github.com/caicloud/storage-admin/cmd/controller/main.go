package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/golang/glog" // -log_dir=./logs -stderrthreshold=info

	"github.com/caicloud/storage-admin/pkg/config"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/controller/storageclass"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

var (
	// k8s about
	kubeHost   string
	kubeConfig string
	// controller
	controllerMaxRetryTimes int

	sigCh  = make(chan os.Signal, 32)
	stopCh = make(chan struct{})
)

func argsInit() {
	flag.StringVar(&kubeHost, config.EnvKubeHost, "", "kube host address")
	flag.StringVar(&kubeConfig, config.EnvKubeConfig, "", "kube config file path")
	flag.IntVar(&controllerMaxRetryTimes, config.EnvControllerMaxRetryTimes, 0, "controller failure max retry time, < 1 means unlimited")
	flag.Set(constants.GLogLogLevelFlagName, config.LogLevel)
	flag.Set(constants.GLogLogVerbosityFlagName, config.LogVerbosity)
	flag.Parse()

	if len(config.KubeHost) > 0 {
		kubeHost = config.KubeHost
	}
	if len(config.KubeConfig) > 0 {
		kubeConfig = config.KubeConfig
	}

	if controllerMaxRetryTimes == 0 {
		controllerMaxRetryTimes = config.ControllerMaxRetryTimes
	}
	glog.Infof("arg %s=%v", config.EnvKubeHost, kubeHost)
	glog.Infof("arg %s=%v", config.EnvKubeConfig, kubeConfig)
	glog.Infof("arg %s=%v", config.EnvControllerMaxRetryTimes, controllerMaxRetryTimes)
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

	// controller
	c, e := storageclass.NewController(kc)
	if e != nil {
		glog.Exitf("NewController failed, %v", e)
	}
	c.SetMaxRetryTimes(controllerMaxRetryTimes)

	signal.Notify(sigCh, os.Kill)
	go func() {
		<-sigCh
		signal.Stop(sigCh)
		glog.Warningf("receive kill signal, stop program")
		close(stopCh)
	}()

	go func() {
		c.Run(1, stopCh)
	}()

	<-stopCh
}
