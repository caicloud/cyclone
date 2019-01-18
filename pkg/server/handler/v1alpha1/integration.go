package v1alpha1

import (
	"context"
	"encoding/json"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	apiv1alpha1 "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
)

// ListIntegrations get integrations the given tenant has access to.
// - ctx Context of the reqeust
// - tenant Tenant
// - pagination Pagination with page and limit.
func ListIntegrations(ctx context.Context, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	// TODO: Need a more efficient way to get paged items.
	secrets, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).List(meta_v1.ListOptions{
		LabelSelector: common.LabelIntegrationType,
	})
	if err != nil {
		log.Errorf("Get integrations from k8s with tenant %s error: %v", tenant, err)
		return nil, err
	}

	items := secrets.Items
	integrations := []apiv1alpha1.Integration{}
	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), integrations), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	for _, secret := range items {
		integration, err := SecretToIntegration(&secret)
		if err != nil {
			continue
		}
		integrations = append(integrations, *integration)
	}

	return types.NewListResponse(int(size), integrations[pagination.Start:end]), nil
}

// SecretToIntegration translates secret to integration
func SecretToIntegration(secret *core_v1.Secret) (*apiv1alpha1.Integration, error) {
	integration := &apiv1alpha1.Integration{}
	err := json.Unmarshal(secret.Data[common.SecretKeyIntegration], integration)
	if err != nil {
		return integration, err
	}

	integration.Metadata.CreationTime = secret.ObjectMeta.CreationTimestamp.String()
	return integration, nil
}

// CreateIntegration creates an integration to store external system info for the tenant.
func CreateIntegration(ctx context.Context, tenant string, in *apiv1alpha1.Integration) (*apiv1alpha1.Integration, error) {
	ns := common.TenantNamespace(tenant)
	secret, err := buildSecret(tenant, in)
	if err != nil {
		return nil, err
	}
	_, err = handler.K8sClient.CoreV1().Secrets(ns).Create(secret)
	if err != nil {
		log.Errorf("Create secret %v for tenant error %v", secret.ObjectMeta.Name, tenant, err)
		return nil, err
	}

	return in, nil
}

func buildSecret(tenant string, in *apiv1alpha1.Integration) (*core_v1.Secret, error) {
	ns := common.TenantNamespace(tenant)
	secretName := common.IntegrationSecret(in.Metadata.Name)

	labels := make(map[string]string)
	labels[common.LabelIntegrationType] = string(in.Spec.Type)

	integration, err := json.Marshal(in)
	if err != nil {
		log.Errorf("Marshal integration %v for tenant %s error %v", in.Metadata.Name, tenant, err)
		return nil, err
	}
	data := make(map[string][]byte)
	data[common.SecretKeyIntegration] = integration

	secret := &core_v1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      secretName,
			Namespace: ns,
			Labels:    labels,
		},
		Data: data,
	}

	return secret, nil
}

// GetIntegration gets an integration with the given name under given tenant.
func GetIntegration(ctx context.Context, tenant, name string) (*apiv1alpha1.Integration, error) {
	secret, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Get(
		common.IntegrationSecret(name), meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return SecretToIntegration(secret)
}

// UpdateIntegration updates an integration with the given tenant name and integration name.
// If updated successfully, return the updated integration.
func UpdateIntegration(ctx context.Context, tenant, name string, in *apiv1alpha1.Integration) (*apiv1alpha1.Integration, error) {
	ns := common.TenantNamespace(tenant)
	secret, err := buildSecret(tenant, in)
	if err != nil {
		return nil, err
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CoreV1().Secrets(ns).Get(
			common.IntegrationSecret(name), meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		newSecret := origin.DeepCopy()
		newSecret.Data = secret.Data
		_, err = handler.K8sClient.CoreV1().Secrets(ns).Update(newSecret)
		return err
	})

	if err != nil {
		return nil, err
	}

	return in, nil
}

// DeleteIntegration deletes a integration with the given tenant and name.
func DeleteIntegration(ctx context.Context, tenant, name string) error {
	return handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Delete(
		common.IntegrationSecret(name), &meta_v1.DeleteOptions{})
}
