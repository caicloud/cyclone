package integration

import (
	"encoding/json"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"

	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
)

// ToSecret converts an integration to k8s secret
func ToSecret(tenant string, in *api.Integration) (*core_v1.Secret, error) {
	objectMeta := in.ObjectMeta
	objectMeta.Name = GetSecretName(in.Name)
	if objectMeta.Labels == nil {
		objectMeta.Labels = make(map[string]string)
	}

	objectMeta.Labels[meta.LabelIntegrationType] = string(in.Spec.Type)
	if in.Spec.Type == api.Cluster && in.Spec.Cluster != nil {
		if in.Spec.Cluster.IsWorkerCluster {
			objectMeta.Labels = meta.AddSchedulableClusterLabel(objectMeta.Labels)
		} else if _, ok := objectMeta.Labels[meta.LabelIntegrationSchedulableCluster]; ok {
			delete(objectMeta.Labels, meta.LabelIntegrationSchedulableCluster)
		}
	}

	integration, err := json.Marshal(in.Spec)
	if err != nil {
		log.Errorf("Marshal integration %v for tenant %s error %v", in.Name, tenant, err)
		return nil, err
	}
	data := make(map[string][]byte)
	data[common.SecretKeyIntegration] = integration

	return &core_v1.Secret{
		ObjectMeta: objectMeta,
		Data:       data,
	}, nil
}

// FromSecret converts a secret to a integration
func FromSecret(secret *core_v1.Secret) (*api.Integration, error) {
	integration := &api.Integration{
		ObjectMeta: secret.ObjectMeta,
	}

	// retrieve integration name
	integration.Name = GetIntegrationName(secret.Name)
	err := json.Unmarshal(secret.Data[common.SecretKeyIntegration], &integration.Spec)
	if err != nil {
		return nil, err
	}

	return integration, nil
}
