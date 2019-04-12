package sorter

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/caicloud/nirvana/log"
	"github.com/thedevsaddam/gojsonq"
	"k8s.io/apimachinery/pkg/runtime"
)

// RuntimeSort is an implementation of the golang sort interface that knows how to sort
// lists of runtime.Object by metadata.creationTimestamp
type RuntimeSort struct {
	field        string
	objs         []runtime.Object
	origPosition []int
	ascending    bool
}

const (
	sortField string = "metadata.creationTimestamp"
)

// NewRuntimeSort creates a sort tool used to sort runtime.Object by metadata.creationTimestamp
func NewRuntimeSort(objs []runtime.Object, ascending bool) *RuntimeSort {
	sorter := &RuntimeSort{field: sortField, objs: objs, origPosition: make([]int, len(objs)), ascending: ascending}

	for ix := range objs {
		sorter.origPosition[ix] = ix
	}

	sort.Sort(sorter)
	return sorter
}

// Len implements interface sort.Interface
func (r *RuntimeSort) Len() int {
	return len(r.objs)
}

// Swap implements interface sort.Interface
func (r *RuntimeSort) Swap(i, j int) {
	r.objs[i], r.objs[j] = r.objs[j], r.objs[i]
	r.origPosition[i], r.origPosition[j] = r.origPosition[j], r.origPosition[i]
}

// Less implements interface sort.Interface
func (r *RuntimeSort) Less(i, j int) bool {
	iObj := r.objs[i]
	jObj := r.objs[j]

	iData, err := json.Marshal(iObj)
	if err != nil {
		log.Errorf("marshl object %#v error:%v", iObj, err)
		return false
	}

	jData, err := json.Marshal(jObj)
	if err != nil {
		log.Errorf("marshl object %#v error:%v", jObj, err)
		return false
	}

	iValue := gojsonq.New().JSONString(string(iData)).Find(r.field)
	jValue := gojsonq.New().JSONString(string(jData)).Find(r.field)

	iTime, err := time.Parse(time.RFC3339, iValue.(string))
	if err != nil {
		log.Errorf("convert to %v Time error", iValue)
		return false
	}
	jTime, err := time.Parse(time.RFC3339, jValue.(string))
	if err != nil {
		log.Errorf("convert to %v Time error", jValue)
		return false
	}

	if r.ascending {
		return iTime.Before(jTime)
	}
	return iTime.After(jTime)
}

// OriginalPosition returns the starting (original) position of a particular index.
// e.g. If OriginalPosition(0) returns 5 than the the item currently at position 0
// was at position 5 in the original unsorted array.
func (r *RuntimeSort) OriginalPosition(ix int) int {
	if ix < 0 || ix > len(r.origPosition) {
		return -1
	}
	return r.origPosition[ix]
}
