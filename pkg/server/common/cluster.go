package common

import (
	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"

	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
)

// NewClusterClient creates a client for k8s cluster
func NewClusterClient(c *api.ClusterCredential, inCluster bool) (*kubernetes.Clientset, error) {
	if inCluster {
		return newInclusterK8sClient()
	}
	return newK8sClient(c)
}

func newK8sClient(c *api.ClusterCredential) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// if KubeConfig is not empty, use it firstly, otherwise, use username/password.
	if c.KubeConfig != nil {
		config, err = clientcmd.NewDefaultClientConfig(*c.KubeConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			log.Infof("NewDefaultClientConfig error: %v", err)
			return nil, err
		}
	} else {
		if c.TLSClientConfig == nil {
			c.TLSClientConfig = &api.TLSClientConfig{Insecure: true}
		}

		config = &rest.Config{
			Host:        c.Server,
			BearerToken: c.BearerToken,
			Username:    c.User,
			Password:    c.Password,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: c.TLSClientConfig.Insecure,
				CAFile:   c.TLSClientConfig.CAFile,
				CAData:   c.TLSClientConfig.CAData,
			},
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return client, nil
}

func newInclusterK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return client, nil
}

// CreateNamespace creates a namespace
func CreateNamespace(name string, client *kubernetes.Clientset) error {
	namespace, err := buildNamespace(name)
	if err != nil {
		log.Warningf("Build namespace %s error %v", name, err)
		return err
	}

	_, err = client.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Info("namespace %s already exists", name)
			return nil
		}

		log.Errorf("Create namespace %s error %v", name, err)
		return err
	}

	return nil
}

func buildNamespace(tenant string) (*core_v1.Namespace, error) {
	// set labels
	label := make(map[string]string)
	label[LabelOwner] = OwnerCyclone

	nsname := TenantNamespace(tenant)
	namespace := &core_v1.Namespace{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:   nsname,
			Labels: label,
		},
	}

	return namespace, nil
}

// CreateResourceQuota creates resource quota for tenant
func CreateResourceQuota(tenant *api.Tenant, namespace string, client *kubernetes.Clientset) error {
	nsname := TenantNamespace(tenant.Metadata.Name)
	if namespace != "" {
		nsname = namespace
	}

	quota, err := buildResourceQuota(tenant)
	if err != nil {
		log.Warningf("Build resource quota for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	_, err = client.CoreV1().ResourceQuotas(nsname).Create(quota)
	if err != nil {
		log.Errorf("Create ResourceQuota for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	return nil
}

func buildResourceQuota(tenant *api.Tenant) (*core_v1.ResourceQuota, error) {
	// parse resource list
	rl, err := ParseResourceList(tenant.Spec.ResourceQuota)
	if err != nil {
		log.Warningf("Parse resource quota for tenant %s error %v", tenant.Metadata.Name, err)
		return nil, err
	}

	quotaName := TenantResourceQuota(tenant.Metadata.Name)
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
	nsname := TenantNamespace(tenant.Metadata.Name)
	if namespace != "" {
		nsname = namespace
	}

	// parse resource list
	rl, err := ParseResourceList(tenant.Spec.ResourceQuota)
	if err != nil {
		log.Warningf("Parse resource quota for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		quota, err := client.CoreV1().ResourceQuotas(nsname).Get(
			TenantResourceQuota(tenant.Metadata.Name), meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("Get ResourceQuota for tenant %s error %v", tenant.Metadata.Name, err)
			return err
		}

		quota.Spec.Hard = rl
		_, err = client.CoreV1().ResourceQuotas(nsname).Update(quota)
		if err != nil {
			log.Errorf("Update ResourceQuota for tenant %s error %v", tenant.Metadata.Name, err)
			return err
		}

		return nil
	})

}

// CreatePVC creates pvc for tenant
func CreatePVC(tenantName, storageClass, size string, namespace string, client *kubernetes.Clientset) error {
	// parse quantity
	resources := make(map[core_v1.ResourceName]resource.Quantity)
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		log.Errorf("Parse Quantity %s error %v", size, err)
		return err
	}
	resources[core_v1.ResourceStorage] = quantity

	// create pvc
	pvcName := TenantPVC(tenantName)
	nsname := TenantNamespace(tenantName)
	if namespace != "" {
		nsname = namespace
	}
	volume := &core_v1.PersistentVolumeClaim{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      pvcName,
			Namespace: nsname,
		},
		Spec: core_v1.PersistentVolumeClaimSpec{
			AccessModes: []core_v1.PersistentVolumeAccessMode{core_v1.ReadWriteMany},
			Resources: core_v1.ResourceRequirements{
				Limits:   resources,
				Requests: resources,
			},
		},
	}

	if storageClass != "" {
		volume.Spec.StorageClassName = &storageClass
	}

	_, err = client.CoreV1().PersistentVolumeClaims(nsname).Create(volume)
	if err != nil {
		log.Errorf("Create persistent volume claim %s error %v", pvcName, err)
		return err
	}

	return nil
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

// UpdatePVC delete the old pvc and recreate another one, so the data of the pvc will lost.
func UpdatePVC(tenantName, storageClass, size string, namespace string, client *kubernetes.Clientset) error {
	// TODO(zhujian7) Need to implement

	return nil
}
