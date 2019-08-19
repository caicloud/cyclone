package usage

import (
	"encoding/json"
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/common"
)

// PVCUsage represents PVC usages in a tenant, values are in human readable format, for example, '8K', '1.2G'.
type PVCUsage struct {
	// Total is total space
	Total string `json:"total"`
	// Used is space used
	Used string `json:"used"`
	// Items are space used by each folder, for example, 'caches' -> '1.2G'
	Items map[string]string `json:"items"`
}

// PVCUsageFloat64 is same to PVCUsage, but have all values in float64 type.
type PVCUsageFloat64 struct {
	// Total is total space
	Total float64
	// Used is space used
	Used float64
	// Items are space used by each folder, for example, 'caches' -> '8096'
	Items map[string]float64
}

// ToFloat64 converts usage values from string to float64.
func (u *PVCUsage) ToFloat64() (*PVCUsageFloat64, error) {
	usage := &PVCUsageFloat64{
		Items: make(map[string]float64),
	}

	total, err := Parse(u.Total)
	if err != nil {
		return nil, fmt.Errorf("parse value %s error: %v", u.Total, err)
	}
	usage.Total = total

	used, err := Parse(u.Used)
	if err != nil {
		return nil, fmt.Errorf("parse value %s error: %v", u.Used, err)
	}
	usage.Used = used

	for k, v := range u.Items {
		fv, err := Parse(v)
		if err != nil {
			return nil, fmt.Errorf("parse value %s error: %v", v, err)
		}
		usage.Items[k] = fv
	}

	return usage, nil
}

// PVCReporter reports PVC usage information.
type PVCReporter interface {
	OverallUsedPercentage() float64
	UsedPercentage(folder string) (float64, error)
}

type pvcReporter struct {
	tenant string
	usage  *PVCUsageFloat64
}

// NewPVCReporter creates a PVC usage reporter.
func NewPVCReporter(client clientset.Interface, tenant string) (PVCReporter, error) {
	if client == nil {
		return nil, fmt.Errorf("k8s client is nil, tenant: %s", tenant)
	}

	u, err := getUsage(client, tenant)
	if err != nil {
		return nil, err
	}

	usage, err := u.ToFloat64()
	if err != nil {
		return nil, fmt.Errorf("convert usage to float64 error: %v", err)
	}

	return &pvcReporter{
		tenant: tenant,
		usage:  usage,
	}, nil
}

// OverallUsedPercentage get overall usage of PVC, for example, 80%.
func (p *pvcReporter) OverallUsedPercentage() float64 {
	if p.usage.Total == 0.0 {
		return 1.0
	}

	return p.usage.Used / p.usage.Total
}

// UsedPercentage gets the percentage a given folder (top level folder) take in the overall PVC.
func (p *pvcReporter) UsedPercentage(folder string) (float64, error) {
	value, ok := p.usage.Items[folder]
	if !ok {
		return 0, fmt.Errorf("no usage info found for folder '%s'", folder)
	}

	if p.usage.Total == 0.0 {
		return 1.0, nil
	}

	return value / p.usage.Total, nil
}

func getUsage(client clientset.Interface, tenant string) (*PVCUsage, error) {
	name := common.TenantNamespace(tenant)
	ns, err := client.CoreV1().Namespaces().Get(common.TenantNamespace(tenant), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get namespace %s for tenant %s error: %v", name, tenant, err)
	}

	raw, ok := ns.Annotations[meta.AnnotationTenantStorageUsage]
	if !ok {
		return nil, fmt.Errorf("annotation %s not foud in namespace %s", meta.AnnotationTenantStorageUsage, name)
	}

	usage := &PVCUsage{}
	if err := json.Unmarshal([]byte(raw), usage); err != nil {
		return nil, fmt.Errorf("unmarshal usage from annotation '%s' error: %v", raw, err)
	}

	return usage, nil
}

var unitMap = map[byte]int64{
	'B': 1,
	'K': 1024,
	'M': 1024 * 1024,
	'G': 1024 * 1024 * 1024,
	'T': 1024 * 1024 * 1024 * 1024,
}

// Parse parses usage value from human readable string to float. For example,
// '8.0K' --> 8.0*1024
// '32M' -> 32*1024*1024
// Value should have format `\d+(\.\d)?[BKMGT]`.
func Parse(value string) (float64, error) {
	if len(value) == 0 {
		return 0, fmt.Errorf("empty value: %s", value)
	}

	if value == "0" {
		return 0, nil
	}

	factor, ok := unitMap[value[len(value)-1]]
	if !ok {
		return 0, fmt.Errorf("invalid value %s, expect end with 'B/K/M/G/T'", value)
	}

	v, err := strconv.ParseFloat(value[:len(value)-1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid value %s, parse to float error: %v", value[:len(value)-1], err)
	}

	return float64(factor) * v, nil
}
