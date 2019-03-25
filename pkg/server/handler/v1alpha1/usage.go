package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/statistic"
	"github.com/caicloud/cyclone/pkg/server/common"
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

		ns.Annotations[common.AnnotationStorageUsage] = string(b)

		_, err = handler.K8sClient.CoreV1().Namespaces().Update(ns)
		if err != nil {
			log.Warningf("Update namespace '%s' error: %v", namespace, err)
		}
		return err
	})
}

// parseUsageData parses the raw usage data, the data is in format: <block-size-in-byte>:<total-blocks>:<free-blocks>
func parseUsageData(data string) (*statistic.Usage, error) {
	parts := strings.Split(data, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid usage string: %s", data)
	}

	blockSize, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid usage string: %s", data)
	}
	total, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid usage string: %s", data)
	}
	free, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid usage string: %s", data)
	}

	return &statistic.Usage{
		Total: total * blockSize,
		Used:  (total - free) * blockSize,
	}, nil
}
