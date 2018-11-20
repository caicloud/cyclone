package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/golang/glog" // -log_dir=./logs -stderrthreshold=info

	"github.com/caicloud/storage-admin/pkg/admission"
	"github.com/caicloud/storage-admin/pkg/config"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

var (
	// k8s about
	kubeHost   string
	kubeConfig string
	// cert
	certNamePath         string
	serverSecretNamePath string
	// resource
	appNamePath string
	// ignore
	ignoredNamespaces string
	ignoredClasses    string
	// http
	port int

	sigCh  = make(chan os.Signal, 32)
	stopCh = make(chan struct{})
)

func argsInit() {
	flag.StringVar(&kubeHost, config.EnvKubeHost, "", "kube host address")
	flag.StringVar(&kubeConfig, config.EnvKubeConfig, "", "kube config file path")
	flag.StringVar(&appNamePath, admission.EnvAppNamePath, admission.DefaultAppNamePath, "app name path, namespace/name")
	flag.StringVar(&certNamePath, admission.EnvCertNamePath, admission.DefaultCertNamePath, "cert name path, namespace/name")
	flag.StringVar(&serverSecretNamePath, admission.EnvServerSecretNamePath, admission.DefaultServerSecretNamePath, "server secret name path, namespace/name")
	flag.StringVar(&ignoredNamespaces, admission.EnvIgnoredNamespaces, admission.DefaultIgnoredNamespaces, "ignored namespaces, split by ','")
	flag.StringVar(&ignoredClasses, admission.EnvIgnoredClasses, admission.DefaultIgnoredClasses, "ignored storage classes, split by ','")
	flag.IntVar(&port, config.EnvListenPort, constants.DefaultListenPort, "admission controller listen port")
	flag.Set(constants.GLogLogLevelFlagName, config.LogLevel)
	flag.Set(constants.GLogLogVerbosityFlagName, config.LogVerbosity)
	flag.Parse()

	if len(config.KubeHost) > 0 {
		kubeHost = config.KubeHost
	}
	if len(config.KubeConfig) > 0 {
		kubeConfig = config.KubeConfig
	}

	if envServerSecretNamePath := os.Getenv(admission.EnvServerSecretNamePath); len(envServerSecretNamePath) > 0 {
		serverSecretNamePath = envServerSecretNamePath
	}

	if envCertNamePath := os.Getenv(admission.EnvCertNamePath); len(envCertNamePath) > 0 {
		certNamePath = envCertNamePath
	}

	if envAppNamePath := os.Getenv(admission.EnvAppNamePath); len(envAppNamePath) > 0 {
		appNamePath = envAppNamePath
	}

	if envIgnoreNamespace := os.Getenv(admission.EnvIgnoredNamespaces); len(envIgnoreNamespace) > 0 {
		ignoredNamespaces = envIgnoreNamespace
	}

	if envIgnoreClasses := os.Getenv(admission.EnvIgnoredClasses); len(envIgnoreClasses) > 0 {
		ignoredClasses = envIgnoreClasses
	}

	if config.ListenPort > 0 {
		port = config.ListenPort
	}

	glog.Infof("arg %s=%v", config.EnvKubeHost, kubeHost)
	glog.Infof("arg %s=%v", config.EnvKubeConfig, kubeConfig)
	glog.Infof("arg %s=%v", admission.EnvAppNamePath, appNamePath)
	glog.Infof("arg %s=%v", admission.EnvCertNamePath, certNamePath)
	glog.Infof("arg %s=%v", admission.EnvServerSecretNamePath, serverSecretNamePath)
	glog.Infof("arg %s=%v", admission.EnvIgnoredNamespaces, ignoredNamespaces)
	glog.Infof("arg %s=%v", admission.EnvIgnoredClasses, ignoredClasses)
	glog.Infof("arg %s=%v", config.EnvListenPort, port)
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

	// admission
	s, e := admission.NewServer(kc, certNamePath, serverSecretNamePath)
	if e != nil {
		glog.Exitf("NewServer failed, %v", e)
	}

	// self registration
	e = s.SelfRegistration(appNamePath)
	if e != nil {
		glog.Exitf("SelfRegistration failed, %v", e)
	}

	// ignore
	if len(ignoredNamespaces) > 0 {
		admission.SetPvcIgnoreNamespaces(ignoredNamespaces)
	}
	if len(ignoredClasses) > 0 {
		admission.SetPvcIgnoreClasses(ignoredClasses)
	}

	signal.Notify(sigCh, os.Kill)
	go func() {
		<-sigCh
		signal.Stop(sigCh)
		glog.Warningf("receive kill signal, stop program")
		close(stopCh)
	}()

	go func() {
		s.Run(stopCh, port)
	}()

	<-stopCh
}
