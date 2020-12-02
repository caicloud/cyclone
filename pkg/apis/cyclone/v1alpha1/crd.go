package v1alpha1

import (
	"context"

	log "github.com/sirupsen/logrus"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func newKubeExtClient(masterURL, kubeConfigPath string) (apiextensionsclient.Interface, error) {
	var config *rest.Config
	var err error
	if kubeConfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags(masterURL, kubeConfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	return apiextensionsclient.NewForConfigOrDie(config), nil
}

// EnsureCRDCreated will create built-in CRDs if they are not exist.
func EnsureCRDCreated(masterURL, kubeConfigPath string) {
	log.Info("start to create crd")
	client, err := newKubeExtClient(masterURL, kubeConfigPath)
	if err != nil {
		log.WithField("error", err).Fatal("new kube ext client error")
	}

	createCRD("resource", "resources", "Resource", []string{"rsc"}, v1beta1.NamespaceScoped, client)
	createCRD("stage", "stages", "Stage", []string{"stg"}, v1beta1.NamespaceScoped, client)
	createCRD("workflow", "workflows", "Workflow", []string{"wf"}, v1beta1.NamespaceScoped, client)
	createCRD("workflowrun", "workflowruns", "WorkflowRun", []string{"wfr"}, v1beta1.NamespaceScoped, client)
	createCRD("workflowtrigger", "workflowtriggers", "WorkflowTrigger", []string{"wft"}, v1beta1.NamespaceScoped, client)
	createCRD("project", "projects", "Project", []string{"proj"}, v1beta1.NamespaceScoped, client)
	createCRD("executioncluster", "executionclusters", "ExecutionCluster", []string{"ec"}, v1beta1.ClusterScoped, client)
}

func createCRD(singular, plural, kind string, shortNames []string, scope v1beta1.ResourceScope, client apiextensionsclient.Interface) {
	crdName := plural + "." + GroupName
	_, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(context.TODO(), crdName, metav1.GetOptions{})
	if err == nil {
		log.WithField("name", crdName).Info("crd already exist")
		return
	}

	if !errors.IsNotFound(err) {
		log.WithField("name", crdName).WithField("error", err).Fatal("check existence of crd error")
		return
	}

	// create crd
	log.WithField("name", crdName).Info("create crd")
	_, err = client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.TODO(), &v1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      crdName,
			Namespace: "default",
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   GroupName,
			Version: Version,
			Scope:   scope,
			Names: v1beta1.CustomResourceDefinitionNames{
				Kind:       kind,
				Plural:     plural,
				Singular:   singular,
				ShortNames: shortNames,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		log.WithField("name", crdName).WithField("error", err).Fatal("create crd error")
		return
	}

}
