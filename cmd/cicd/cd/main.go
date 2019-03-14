package main

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/caicloud/cyclone/pkg/cicd/cd"
)

func main() {
	configure, err := cd.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("To update %s: '%s/%s'", configure.Deployment.Type, configure.Deployment.Namespace, configure.Deployment.Name)
	for _, u := range configure.Images {
		log.Infof("Container: %s --> %s", u.Container, u.Image)
	}

	client := getClient("")
	err = cd.UpdateDeployment(client, configure)
	if err != nil {
		log.Errorf("%v", err)
		log.Fatal(err)
	}
}

func getClient(kubeConfigPath string) kubernetes.Interface {
	var config *rest.Config
	var err error
	if kubeConfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			log.Fatalf("create config error: %v", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("create config error: %v", err)
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("create client error: %v", err)
	}

	return client
}
