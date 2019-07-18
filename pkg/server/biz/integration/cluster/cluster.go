package cluster

import (
	"github.com/caicloud/nirvana/log"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/integration"
	"github.com/caicloud/cyclone/pkg/server/biz/pvc"
	"github.com/caicloud/cyclone/pkg/server/biz/tenant"
	svrcommon "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// Open opens cluster to run workload.
func Open(client clientset.Interface, in *api.Integration, tenantName string) (err error) {
	// Convert the returned error if it is a k8s error.
	defer func() {
		err = cerr.ConvertK8sError(err)
	}()

	cluster := in.Spec.Cluster
	clusterClient, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
	if err != nil {
		log.Errorf("new cluster client for tenant %s error %v", tenantName, err)
		return
	}

	tenant, err := tenant.Get(client, tenantName)
	if err != nil {
		log.Errorf("Get tenant %s info error %v", tenantName, err)
		return
	}

	// Create namespace if not exist.
	if cluster.Namespace != "" && cluster.Namespace != svrcommon.TenantNamespace(tenant.Name) {
		// Check if namespace exist, if not found or failed to get, return error
		_, err = clusterClient.CoreV1().Namespaces().Get(cluster.Namespace, meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("Get namespace %s error %v", cluster.Namespace, err)
			return
		}
		log.Infof("Namespace %s already existed, no need to create it", cluster.Namespace)
	} else {
		cluster.Namespace = svrcommon.TenantNamespace(tenant.Name)
		err = svrcommon.CreateNamespace(tenant.Name, clusterClient)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Errorf("Create namespace for tenant %s error %v", tenantName, err)
				return
			}
			log.Infof("Namespace %s already exist", cluster.Namespace)
		}
	}

	// Create default service account for the namespace
	if err = EnsureServiceAccount(clusterClient, cluster.Namespace); err != nil {
		log.Errorf("Create service account error: %v", err)
		return
	}

	// Create resource quota
	err = svrcommon.CreateResourceQuota(tenant, cluster.Namespace, clusterClient)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Errorf("Create resource quota for tenant %s error %v", tenantName, err)
			return
		}
		log.Infof("Resource quota for tenant %s already exist", tenantName)
	}

	// If a PVC has been configured, check existence of it.
	if cluster.PVC != "" && cluster.PVC != svrcommon.TenantPVC(tenant.Name) {
		_, err = clusterClient.CoreV1().PersistentVolumeClaims(cluster.Namespace).Get(cluster.PVC, meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("Get pvc %s error %v", cluster.PVC, err)
			return
		}
	} else {
		if tenant.Spec.PersistentVolumeClaim.Size == "" {
			tenant.Spec.PersistentVolumeClaim.Size = config.Config.DefaultPVCConfig.Size
		}
		if tenant.Spec.PersistentVolumeClaim.StorageClass == "" {
			tenant.Spec.PersistentVolumeClaim.StorageClass = config.Config.DefaultPVCConfig.StorageClass
		}

		err = pvc.CreatePVC(tenant.Name, tenant.Spec.PersistentVolumeClaim.StorageClass,
			tenant.Spec.PersistentVolumeClaim.Size, cluster.Namespace, clusterClient)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Errorf("create pvc for tenant %s error %v", tenantName, err)
				return
			}
			log.Infof("PVC for tenant %s already exist", tenantName)
		}

		cluster.PVC = svrcommon.TenantPVC(tenant.Name)
	}

	clusterName := in.Spec.Cluster.ClusterName
	if in.Spec.Cluster.IsControlCluster {
		clusterName = common.ControlClusterName
	}

	// Create ExecutionCluster resource for Workflow Engine to use
	_, err = client.CycloneV1alpha1().ExecutionClusters().Create(&v1alpha1.ExecutionCluster{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.ExecutionClusterSpec{
			Credential: cluster.Credential,
		},
	})

	// Execution cluster is system level, so different tenants may try to create ExecutionCluster CR for the same
	// cluster, so if the CR already exists, just ignore it.
	if err != nil && errors.IsAlreadyExists(err) {
		log.Infof("ExecutionCluster resource for %s already exist", in.Name)
		return nil
	}

	return err
}

// Close closes a cluster
func Close(client clientset.Interface, in *api.Integration, tenant string) (err error) {
	// Convert the returned error if it is a k8s error.
	defer func() {
		err = cerr.ConvertK8sError(err)
	}()
	cluster := in.Spec.Cluster

	//// new cluster clientset
	//clusterClient, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
	//if err != nil {
	//	log.Errorf("new cluster client error %v", err)
	//	return
	//}

	// delete namespace which is created by cyclone
	if cluster.Namespace == svrcommon.TenantNamespace(tenant) {
		// if is a user cluster and namespace are created by cyclone, delete the namespace directly
		if !cluster.IsControlCluster && cluster.Namespace == svrcommon.TenantNamespace(tenant) {
			err = client.CoreV1().Namespaces().Delete(cluster.Namespace, &meta_v1.DeleteOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					log.Errorf("delete namespace %s error %v", cluster.Namespace, err)
					return
				}
				log.Warningf("namespace %s not found", cluster.Namespace)
				err = nil
			}

			// if namespace is deleted, will exit, no need delete others resources.
			return
		}
	}

	// delete resource quota
	quotaName := svrcommon.TenantResourceQuota(tenant)
	err = client.CoreV1().ResourceQuotas(cluster.Namespace).Delete(quotaName, &meta_v1.DeleteOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorf("delete resource quota %s error %v", quotaName, err)
			return
		}
		log.Warningf("resource quota %s not found", quotaName)
		err = nil
	}

	// Delete the PVC watcher deployment.
	//err = usage.DeletePVCUsageWatcher(clusterClient, cluster.Namespace)
	//if err != nil {
	//	log.Warningf("Delete PVC watcher '%s' error: %v", usage.PVCWatcherName, err)
	//}

	// delete pvc which is created by cyclone
	if cluster.PVC == svrcommon.TenantPVC(tenant) {
		err = client.CoreV1().PersistentVolumeClaims(cluster.Namespace).Delete(cluster.PVC, &meta_v1.DeleteOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				log.Errorf("delete pvc %s error %v", cluster.PVC, err)
				return
			}
			log.Warningf("pvc %s not found", cluster.PVC)
			err = nil
		}
		return
	}

	return nil
}

// UpdateClusterIntegration ...
func UpdateClusterIntegration(client clientset.Interface, tenant, name string, in *api.Integration) error {
	secret, err := client.CoreV1().Secrets(svrcommon.TenantNamespace(tenant)).Get(integration.GetSecretName(name), meta_v1.GetOptions{})
	if err != nil {
		return nil
	}

	oldIn, err := integration.FromSecret(secret)
	if err != nil {
		log.Errorf("Get integration %s error %v", name, err)
		return err
	}

	oldCluster := oldIn.Spec.Cluster
	cluster := in.Spec.Cluster
	if cluster.Namespace == "" {
		cluster.Namespace = oldCluster.Namespace
	} else if cluster.Namespace == oldCluster.Namespace && cluster.PVC == "" {
		cluster.PVC = oldCluster.PVC
	}
	// Turn on worker cluster
	if !oldCluster.IsWorkerCluster && cluster.IsWorkerCluster {
		// open cluster for the tenant, create namespace and pvc
		err := Open(client, in, tenant)
		if err != nil {
			return err
		}
	}

	// Turn off worker cluster
	if oldCluster.IsWorkerCluster && !cluster.IsWorkerCluster {
		// close cluster for the tenant, delete namespace
		err := Close(client, oldIn, tenant)
		if err != nil {
			return err
		}
	}

	// TODO(zhujian7): namespace or pvc changed
	if oldCluster.IsWorkerCluster && cluster.IsWorkerCluster {
		if oldCluster.Namespace != cluster.Namespace {
			log.Info("can not process updating namespace, namespace changed from %s to %s", oldCluster.Namespace, cluster.Namespace)
		}

		if oldCluster.PVC != cluster.PVC {
			log.Info("can not process updating pvc, pvc changed from %s to %s", oldCluster.PVC, cluster.PVC)
		}
	}

	return nil
}

// GetSchedulableClusters gets all clusters which are used to perform workload for this tenant.
func GetSchedulableClusters(client clientset.Interface, tenant string) ([]api.Integration, error) {
	secrets, err := client.CoreV1().Secrets(svrcommon.TenantNamespace(tenant)).List(meta_v1.ListOptions{
		LabelSelector: meta.SchedulableClusterSelector(),
	})
	if err != nil {
		log.Errorf("Get integrations from k8s with tenant %s error: %v", tenant, err)
		return nil, cerr.ConvertK8sError(err)
	}

	integrations := []api.Integration{}
	for _, secret := range secrets.Items {
		integration, err := integration.FromSecret(&secret)
		if err != nil {
			continue
		}
		integrations = append(integrations, *integration)
	}

	return integrations, nil
}
