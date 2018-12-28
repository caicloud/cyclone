package v1alpha1

import (
	log "github.com/sirupsen/logrus"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func newKubeExtClient(masterUrl, kubeConfigPath string) (apiextensionsclient.Interface, error) {
	var config *rest.Config
	var err error
	if kubeConfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags(masterUrl, kubeConfigPath)
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
func EnsureCRDCreated(masterUrl, kubeConfigPath string) {
	log.Info("start to creat crd")
	client, err := newKubeExtClient(masterUrl, kubeConfigPath)
	if err != nil {
		log.WithField("error", err).Fatal("new kube ext client error")
	}

	createCRD("resource", "resources", "Resource", []string{"rsc"}, client)
	createCRD("stage", "stages", "Stage", []string{"stg"}, client)
	createCRD("workflow", "workflows", "Workflow", []string{"wf"}, client)
	createCRD("workflowrun", "workflowruns", "WorkflowRun", []string{"wfr"}, client)
	createCRD("stagetemplate", "stagetemplates", "StageTemplate", []string{"stpl"}, client)
	createCRD("workflowparam", "workflowparams", "WorkflowParams", []string{"wfp"}, client)
	createCRD("workflowtrigger", "workflowtriggers", "WorkflowTrigger", []string{"wft"}, client)
}

func createCRD(singular, plural, kind string, shortNames []string, client apiextensionsclient.Interface) {
	crdName := plural + "." + GroupName
	_, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
	if err != nil {
		if err.(*errors.StatusError).Status().Reason == metav1.StatusReasonNotFound {
			log.WithField("name", crdName).Info("create crd")
			_, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&v1beta1.CustomResourceDefinition{
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
					Scope:   v1beta1.NamespaceScoped,
					Names: v1beta1.CustomResourceDefinitionNames{
						Kind:       kind,
						Plural:     plural,
						Singular:   singular,
						ShortNames: shortNames,
					},
				},
			})
			if err != nil {
				log.WithField("name", crdName).WithField("error", err).Fatal("create crd error")
			}
		} else {
			log.WithField("name", crdName).WithField("error", err).Fatal("check existence of crd error")
		}
	} else {
		log.WithField("name", crdName).Info("crd already exist")
	}
}
