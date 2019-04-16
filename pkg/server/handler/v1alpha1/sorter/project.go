package sorter

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

// ProjectSort implements sort.Interface{}
type ProjectSort struct {
	objs      []v1alpha1.Project
	ascending bool
}

// NewProjectSorter ...
func NewProjectSorter(objs []v1alpha1.Project, ascending bool) *ProjectSort {
	return &ProjectSort{
		objs:      objs,
		ascending: ascending,
	}
}

// Len is the number of elements in the collection.
func (s *ProjectSort) Len() int {
	return len(s.objs)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s *ProjectSort) Less(i, j int) bool {
	if s.ascending {
		return s.objs[i].CreationTimestamp.Before(&s.objs[j].CreationTimestamp)
	}
	return s.objs[i].CreationTimestamp.After(s.objs[j].CreationTimestamp.Time)
}

// Swap swaps the elements with indexes i and j.
func (s *ProjectSort) Swap(i, j int) {
	s.objs[i], s.objs[j] = s.objs[j], s.objs[i]
}
