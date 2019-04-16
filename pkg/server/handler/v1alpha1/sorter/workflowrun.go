package sorter

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

// WorkflowRunSort implements the sort.Interface{}
type WorkflowRunSort struct {
	objs      []v1alpha1.WorkflowRun
	ascending bool
}

// NewWorkflowRunSorter ...
func NewWorkflowRunSorter(objs []v1alpha1.WorkflowRun, ascending bool) *WorkflowRunSort {
	return &WorkflowRunSort{
		objs:      objs,
		ascending: ascending,
	}
}

// Len is the number of elements in the collection.
func (s *WorkflowRunSort) Len() int {
	return len(s.objs)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s *WorkflowRunSort) Less(i, j int) bool {
	if s.ascending {
		return s.objs[i].CreationTimestamp.Before(&s.objs[j].CreationTimestamp)
	}
	return s.objs[i].CreationTimestamp.After(s.objs[j].CreationTimestamp.Time)
}

// Swap swaps the elements with indexes i and j.
func (s *WorkflowRunSort) Swap(i, j int) {
	s.objs[i], s.objs[j] = s.objs[j], s.objs[i]
}
