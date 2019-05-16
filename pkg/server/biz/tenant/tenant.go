package tenant

import (
	"encoding/json"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/utils"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// Get ...
func Get(client clientset.Interface, name string) (*api.Tenant, error) {
	namespace, err := client.CoreV1().Namespaces().Get(common.TenantNamespace(name), meta_v1.GetOptions{})
	if err != nil {
		log.Errorf("Get namespace for tenant %s error %v", name, err)
		return nil, cerr.ConvertK8sError(err)
	}

	return FromNamespace(namespace)
}

// FromNamespace converts a namespace to a tenant
func FromNamespace(namespace *core_v1.Namespace) (*api.Tenant, error) {
	tenant := &api.Tenant{
		ObjectMeta: namespace.ObjectMeta,
	}

	tenant.Name = common.NamespaceTenant(namespace.Name)
	annotationTenant := namespace.Annotations[meta.AnnotationTenantInfo]
	err := json.Unmarshal([]byte(annotationTenant), &tenant.Spec)
	if err != nil {
		log.Errorf("Unmarshal tenant annotation %s error %v", annotationTenant, err)
		return tenant, err
	}

	// Delete tenant info annotation to keep clean
	delete(tenant.Annotations, meta.AnnotationTenantInfo)
	return tenant, nil
}

// CreateNamespace ...
func CreateNamespace(client clientset.Interface, tenant *api.Tenant) error {
	objectMeta := tenant.ObjectMeta

	// build namespace name
	objectMeta.Name = common.TenantNamespace(tenant.Name)

	// marshal tenant and set it into namespace annotation
	b, err := json.Marshal(tenant.Spec)
	if err != nil {
		log.Warningf("Marshal tenant %s error %v", tenant.Name, err)
		return err
	}
	if objectMeta.Annotations == nil {
		objectMeta.Annotations = make(map[string]string)
	}
	objectMeta.Annotations[meta.AnnotationTenantInfo] = string(b)

	// set labels
	if objectMeta.Labels == nil {
		objectMeta.Labels = make(map[string]string)
	}
	objectMeta.Labels[meta.LabelTenantName] = tenant.Name

	_, err = client.CoreV1().Namespaces().Create(&core_v1.Namespace{
		ObjectMeta: objectMeta,
	})
	if err != nil {
		log.Warningf("Create namespace for tenant %s error %v", tenant.Name, err)
		if errors.IsAlreadyExists(err) {
			tenant.Labels = objectMeta.Labels
			return UpdateNamespace(client, tenant)
		}
		return cerr.ConvertK8sError(err)
	}

	return nil
}

// UpdateNamespace ...
func UpdateNamespace(client clientset.Interface, tenant *api.Tenant) error {
	t, err := json.Marshal(tenant.Spec)
	if err != nil {
		log.Warningf("Marshal tenant %s error %v", tenant.Name, err)
		return err
	}

	// update namespace annotation with retry
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := client.CoreV1().Namespaces().Get(common.TenantNamespace(tenant.Name), meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("Get namespace for tenant %s error %v", tenant.Name, err)
			return cerr.ConvertK8sError(err)
		}

		newNs := origin.DeepCopy()
		newNs.Annotations = utils.MergeMap(tenant.Annotations, newNs.Annotations)
		newNs.Labels = utils.MergeMap(tenant.Labels, newNs.Labels)
		newNs.Annotations[meta.AnnotationTenantInfo] = string(t)

		_, err = client.CoreV1().Namespaces().Update(newNs)
		if err != nil {
			log.Errorf("Update namespace for tenant %s error %v", tenant.Name, err)
			return cerr.ConvertK8sError(err)
		}
		return nil
	})

}
