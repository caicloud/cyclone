package sorter

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

// WorkflowTriggerSort implements the sort.Interface{}
type WorkflowTriggerSort struct {
	objs      []v1alpha1.WorkflowTrigger
	ascending bool
}

// NewWorkflowTriggerSorter ...
func NewWorkflowTriggerSorter(objs []v1alpha1.WorkflowTrigger, ascending bool) *WorkflowTriggerSort {
	return &WorkflowTriggerSort{
		objs:      objs,
		ascending: ascending,
	}
}

// Len is the number of elements in the collection.
func (s *WorkflowTriggerSort) Len() int {
	return len(s.objs)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s *WorkflowTriggerSort) Less(i, j int) bool {
	if s.ascending {
		return s.objs[i].CreationTimestamp.Before(&s.objs[j].CreationTimestamp)
	}
	return s.objs[i].CreationTimestamp.After(s.objs[j].CreationTimestamp.Time)
}

// Swap swaps the elements with indexes i and j.
func (s *WorkflowTriggerSort) Swap(i, j int) {
	s.objs[i], s.objs[j] = s.objs[j], s.objs[i]
}
