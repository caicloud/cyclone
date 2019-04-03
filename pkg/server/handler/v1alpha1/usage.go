package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/handler"
)

// ReportStorageUsage reports storage usage of a namespace.
func ReportStorageUsage(ctx context.Context, namespace string, request v1alpha1.StorageUsage) error {
	log.Infof("update pvc storage usage, namespace: %s, usage: %s/%s", namespace, request.Used, request.Total)
	b, err := json.Marshal(request)
	if err != nil {
		log.Warningf("Marshal usage error: %v", err)
		return fmt.Errorf("marshal usage error: %v", err)
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ns, err := handler.K8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if err != nil {
			log.Errorf("Get namespace '%s' error: %v", namespace, err)
			return err
		}

		if ns.Annotations == nil {
			ns.Annotations = make(map[string]string)
		}

		ns.Annotations[meta.AnnotationTenantStorageUsage] = string(b)

		_, err = handler.K8sClient.CoreV1().Namespaces().Update(ns)
		if err != nil {
			log.Warningf("Update namespace '%s' error: %v", namespace, err)
		}
		return err
	})
}
