/*
Copyright 2016 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"

	"k8s.io/client-go/pkg/api/resource"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

const (
	// ResourceCPU in cores. (500m = .5 cores)
	ResourceCPU = "cpu"
	// ResourceMemory in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceMemory = "memory"
	// ResourceRequestsCPU CPU request, in cores. (500m = .5 cores)
	ResourceRequestsCPU = "requests.cpu"
	// ResourceRequestsMemory Memory request, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceRequestsMemory = "requests.memory"
	// ResourceLimitsCPU CPU limit, in cores. (500m = .5 cores)
	ResourceLimitsCPU = "limits.cpu"
	// ResourceLimitsMemory Memory limit, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceLimitsMemory = "limits.memory"
)

var (
	// ZeroQuantity ...
	ZeroQuantity = NewDecimalQuantity(0)
	// ZeroQuota ...
	ZeroQuota = Quota{
		// ResourceCPU:            ZeroQuantity,
		// ResourceMemory:         ZeroQuantity,
		// ResourceRequestsCPU:    ZeroQuantity,
		// ResourceRequestsMemory: ZeroQuantity,
		ResourceLimitsCPU:    ZeroQuantity,
		ResourceLimitsMemory: ZeroQuantity,
	}

	// DefaultLimitCPU 500m = 0.5 core = 500 * 100 * 100
	DefaultLimitCPU = MustParseCPU(0.5)
	// DefaultLimitMemory 500Mi = 500MiB = 500 * 1024 * 1024
	DefaultLimitMemory = NewBinaryQuantity(500 * 1024 * 1024)
	// DefaultQuota ...
	DefaultQuota = Quota{
		ResourceLimitsCPU:    DefaultLimitCPU,
		ResourceLimitsMemory: DefaultLimitMemory,
	}
)

// ------------------------------------------------------------------

// Quantity warps resource.Quantity to implement cli.Generic interface
type Quantity struct {
	resource.Quantity
}

// MustParseCPU turns the given float(in cores, such as 1.5 cores) into a quantity or panics;
// for tests or others cases where you know the value is valid.
func MustParseCPU(value float64) *Quantity {
	return NewQuantityFor(resource.MustParse(fmt.Sprintf("%f", value)))
}

// MustParseMemory turns the given float(in Bytes, such as 500*1024*1024 bytes) into a quantity or panics;
// for tests or others cases where you know the value is valid.
func MustParseMemory(value float64) *Quantity {
	return NewQuantityFor(resource.MustParse(BytesSize(value)))
}

// NewDecimalQuantity creates a new Quantity with resource.DecimalSI Format
func NewDecimalQuantity(value int) *Quantity {
	return &Quantity{*resource.NewQuantity(int64(value), resource.DecimalSI)}
}

// NewBinaryQuantity creates a new Quantity with resource.BinarySI Format
func NewBinaryQuantity(value int) *Quantity {
	return &Quantity{*resource.NewQuantity(int64(value), resource.BinarySI)}
}

// NewQuantity creates a new Quantity
func NewQuantity(value int64, format resource.Format) *Quantity {
	return &Quantity{*resource.NewQuantity(value, format)}
}

// NewQuantityFor reates a new Quantity from resource.Quantity
func NewQuantityFor(q resource.Quantity) *Quantity {
	return &Quantity{q}
}

// DeepCopy returns a deep-copy of the Quantity value.  Note that the method
// receiver is a value, so we can mutate it in-place and return it.
func (q Quantity) DeepCopy() Quantity {
	dq := q.Quantity.DeepCopy()
	q.Quantity = dq
	return q
}

// Set implements cli.Generic interface
func (q *Quantity) Set(value string) error {
	if len(value) == 0 {
		value = "0"
	}

	rq, err := resource.ParseQuantity(value)
	if err != nil {
		return err
	}

	q.Quantity = rq
	return nil
}

func (q Quantity) String() string {
	return q.Quantity.String()
}

// -----------------------------------------------------------------

// Quota ...
type Quota map[string]*Quantity

// IsZero returns true if the all Quantities in Quota are equal to zero.
func (q Quota) IsZero() bool {
	for _, v := range q {
		if !v.IsZero() {
			return false
		}
	}
	return true
}

// Add adds the provide y Quota to the current value.
func (q Quota) Add(y Quota) {
	for k, v := range q {
		vy, ok := y[k]
		if !ok {
			vy = ZeroQuantity
		}
		v.Add(vy.Quantity)
	}
}

// Sub subtracts the provided y Quota from the current value in place.
func (q Quota) Sub(y Quota) {
	for k, v := range q {
		vy, ok := y[k]
		if !ok {
			vy = ZeroQuantity
		}
		v.Sub(vy.Quantity)
	}
}

// Enough returns true if the Quota is greater than y plus z.
func (q Quota) Enough(y Quota, z Quota) bool {
	// deepcopy q because the following step will change it
	qq := q.DeepCopy()

	for k, v := range qq {
		vy, ok := y[k]
		if !ok {
			vy = ZeroQuantity
		}
		vz, ok := z[k]
		if !ok {
			vz = ZeroQuantity
		}
		v.Sub(vy.Quantity)
		cmp := v.Cmp(vz.Quantity)
		if cmp < 0 {
			return false
		}
	}
	return true
}

// DeepCopy returns a deep-copy of the Quota value.
func (q Quota) DeepCopy() Quota {
	ret := make(Quota, len(q))
	for k, v := range q {
		vv := v.DeepCopy()
		ret[k] = &vv
	}
	return ret
}

// ToK8SQuota converts Quota to k8s resource quota type
func (q Quota) ToK8SQuota() apiv1.ResourceRequirements {
	rr := apiv1.ResourceRequirements{
		Limits:   make(apiv1.ResourceList),
		Requests: make(apiv1.ResourceList),
	}

	for k, v := range q {
		keys := strings.Split(k, ".")
		n := len(keys)
		switch {
		case keys[0] == "request":
			rr.Requests[apiv1.ResourceName(keys[1])] = v.Quantity
		default:
			key := keys[0]
			if n == 2 {
				// limit.cpu
				key = keys[1]
			}
			rr.Limits[apiv1.ResourceName(key)] = v.Quantity
		}
	}

	return rr
}

// ToDockerQuota converts Quota to docker resource quota type
func (q Quota) ToDockerQuota() container.Resources {
	q.SetDefault()
	return container.Resources{
		NanoCPUs: q[ResourceLimitsCPU].ScaledValue(-9), // NanoCPUs
		Memory:   q[ResourceLimitsMemory].Value(),      // in bytes
	}
}

// SetDefault fills quota with default quantity
func (q Quota) SetDefault() {
	if _, ok := q[ResourceLimitsCPU]; !ok {
		q[ResourceLimitsCPU] = DefaultQuota[ResourceLimitsCPU]
	}
	if _, ok := q[ResourceLimitsMemory]; !ok {
		q[ResourceLimitsMemory] = DefaultQuota[ResourceLimitsMemory]
	}
}

// ---------------------------------------------------

// Resource describes cloud resource include limit and used quota
type Resource struct {
	Limit Quota `json:"limit,omitempty" bson:"limit,omitempty"`
	Used  Quota `json:"used,omitempty" bson:"used,omitempty"`
}

// NewResource returns a new Resource
func NewResource() *Resource {
	return &Resource{
		Limit: ZeroQuota.DeepCopy(),
		Used:  ZeroQuota.DeepCopy(),
	}
}

// Add adds the provided Resource y to current Resource
func (r *Resource) Add(y *Resource) {
	r.Limit.Add(y.Limit)
	r.Used.Add(y.Used)
}

// ---------------------------------------------------

// start from KiB because resource.Quantity can only parse from KiB
var binaryAbbrs = []string{"Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi", "Yi"}

func getSizeAndUnit(size float64, base float64, _map []string) (float64, string) {
	i := 0
	unitsLimit := len(_map) - 1
	for size >= base && i < unitsLimit {
		size = size / base
		i++
	}
	return size, _map[i]
}

// CustomSize returns a human-readable approximation of a size
// using custom format.
func CustomSize(format string, size float64, base float64, _map []string) string {
	size, unit := getSizeAndUnit(size, base, _map)
	return fmt.Sprintf(format, size, unit)
}

// BytesSize returns a human-readable size in bytes, kibibytes,
// mebibytes, gibibytes, or tebibytes (eg. "44kiB", "17MiB").
func BytesSize(size float64) string {
	// resource.Quantity can not parse n Bytes, it starts from KiB
	return CustomSize("%.4g%s", size/1024.0, 1024.0, binaryAbbrs)
}
