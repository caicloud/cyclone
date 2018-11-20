package kubernetes

import (
	"strings"

	apiextensions "github.com/caicloud/clientset/pkg/apis/apiextensions/v1beta1"
	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CrdKindNameStorageType    = "StorageType"
	CrdKindNameStorageService = "StorageService"
)

func NewStorageTypeCrd() *apiextensions.CustomResourceDefinition {
	crd, _ := NewClusterScopedCrd(resv1b1.DefaultCrdResourceGroup, resv1b1.DefaultApiVersion, CrdKindNameStorageType, "")
	return crd
}
func NewStorageServiceCrd() *apiextensions.CustomResourceDefinition {
	crd, _ := NewClusterScopedCrd(resv1b1.DefaultCrdResourceGroup, resv1b1.DefaultApiVersion, CrdKindNameStorageService, "")
	return crd
}

func NewClusterScopedCrd(group, version, kind, plural string) (*apiextensions.CustomResourceDefinition, error) {
	singular := strings.ToLower(kind)
	if len(plural) == 0 {
		if strings.HasSuffix(singular, "s") {
			plural = singular + "es"
		} else {
			plural = singular + "s"
		}
	}
	// TODO check parameters
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: plural + "." + group,
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   group,
			Version: version,
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:   plural,
				Singular: singular,
				Kind:     kind,
				ListKind: kind + "List",
			},
			Scope: apiextensions.ClusterScoped,
		},
	}, nil
}

func InitStorageAdminMetaMainCluster(cs Interface) error {
	crds := []*apiextensions.CustomResourceDefinition{
		NewStorageTypeCrd(),
		NewStorageServiceCrd(),
	}
	for _, crd := range crds {
		_, e := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
		if e == nil {
			continue
		}
		if !IsNotFound(e) {
			return e
		}
		_, e = cs.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
		if e != nil && !IsAlreadyExists(e) {
			return e
		}
	}
	return nil
}
