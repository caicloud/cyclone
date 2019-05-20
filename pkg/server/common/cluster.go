package common

import (
	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
)

// CreateNamespace creates a namespace
func CreateNamespace(tenant string, client *kubernetes.Clientset) error {
	namespace := buildNamespace(tenant)

	_, err := client.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Infof("namespace %s already exists", namespace.Name)
			return nil
		}

		log.Errorf("Create namespace %s error %v", namespace.Name, err)
		return err
	}

	return nil
}

func buildNamespace(tenant string) *core_v1.Namespace {
	nsname := TenantNamespace(tenant)
	return &core_v1.Namespace{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: nsname,
			Labels: map[string]string{
				meta.LabelTenantName: tenant,
			},
		},
	}
}

// CreateResourceQuota creates resource quota for tenant
func CreateResourceQuota(tenant *api.Tenant, namespace string, client *kubernetes.Clientset) error {
	nsname := TenantNamespace(tenant.Name)
	if namespace != "" {
		nsname = namespace
	}

	quota, err := buildResourceQuota(tenant)
	if err != nil {
		log.Warningf("Build resource quota for tenant %s error %v", tenant.Name, err)
		return err
	}

	_, err = client.CoreV1().ResourceQuotas(nsname).Create(quota)
	if err != nil {
		log.Errorf("Create ResourceQuota for tenant %s error %v", tenant.Name, err)
		return err
	}

	return nil
}

func buildResourceQuota(tenant *api.Tenant) (*core_v1.ResourceQuota, error) {
	// parse resource list
	rl, err := ParseResourceList(tenant.Spec.ResourceQuota)
	if err != nil {
		log.Warningf("Parse resource quota for tenant %s error %v", tenant.Name, err)
		return nil, err
	}

	quotaName := TenantResourceQuota(tenant.Name)
	quota := &core_v1.ResourceQuota{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: quotaName,
		},
		Spec: core_v1.ResourceQuotaSpec{
			Hard: rl,
		},
	}

	return quota, nil
}

// UpdateResourceQuota updates resource quota for tenant
func UpdateResourceQuota(tenant *api.Tenant, namespace string, client *kubernetes.Clientset) error {
	nsname := TenantNamespace(tenant.Name)
	if namespace != "" {
		nsname = namespace
	}

	// parse resource list
	rl, err := ParseResourceList(tenant.Spec.ResourceQuota)
	if err != nil {
		log.Warningf("Parse resource quota for tenant %s error %v", tenant.Name, err)
		return err
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		quota, err := client.CoreV1().ResourceQuotas(nsname).Get(
			TenantResourceQuota(tenant.Name), meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("Get ResourceQuota for tenant %s error %v", tenant.Name, err)
			return err
		}

		quota.Spec.Hard = rl
		_, err = client.CoreV1().ResourceQuotas(nsname).Update(quota)
		if err != nil {
			log.Errorf("Update ResourceQuota for tenant %s error %v", tenant.Name, err)
			return err
		}

		return nil
	})

}

// ParseResourceList parse resouces from 'map[string]string' to 'ResourceList'
func ParseResourceList(resources map[core_v1.ResourceName]string) (map[core_v1.ResourceName]resource.Quantity, error) {
	rl := make(map[core_v1.ResourceName]resource.Quantity)

	for r, q := range resources {
		quantity, err := resource.ParseQuantity(q)
		if err != nil {
			log.Errorf("Parse %s Quantity %s error %v", r, q, err)
			return nil, err
		}
		rl[r] = quantity
	}

	return rl, nil
}
