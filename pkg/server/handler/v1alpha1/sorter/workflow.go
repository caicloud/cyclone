package sorter

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

// WorkflowSort implements the sort.Interface{}
type WorkflowSort struct {
	objs      []v1alpha1.Workflow
	ascending bool
}

// NewWorkflowSorter ...
func NewWorkflowSorter(objs []v1alpha1.Workflow, ascending bool) *WorkflowSort {
	return &WorkflowSort{
		objs:      objs,
		ascending: ascending,
	}
}

// Len is the number of elements in the collection.
func (s *WorkflowSort) Len() int {
	return len(s.objs)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s *WorkflowSort) Less(i, j int) bool {
	if s.ascending {
		return s.objs[i].CreationTimestamp.Before(&s.objs[j].CreationTimestamp)
	}
	return s.objs[i].CreationTimestamp.After(s.objs[j].CreationTimestamp.Time)
}

// Swap swaps the elements with indexes i and j.
func (s *WorkflowSort) Swap(i, j int) {
	s.objs[i], s.objs[j] = s.objs[j], s.objs[i]
}
