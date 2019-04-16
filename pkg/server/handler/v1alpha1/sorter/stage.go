package sorter

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

// StageSort implements the sort.Interface{}
type StageSort struct {
	objs      []v1alpha1.Stage
	ascending bool
}

// NewStageSorter ...
func NewStageSorter(objs []v1alpha1.Stage, ascending bool) *StageSort {
	return &StageSort{
		objs:      objs,
		ascending: ascending,
	}
}

// Len is the number of elements in the collection.
func (s *StageSort) Len() int {
	return len(s.objs)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s *StageSort) Less(i, j int) bool {
	if s.ascending {
		return s.objs[i].CreationTimestamp.Before(&s.objs[j].CreationTimestamp)
	}
	return s.objs[i].CreationTimestamp.After(s.objs[j].CreationTimestamp.Time)
}

// Swap swaps the elements with indexes i and j.
func (s *StageSort) Swap(i, j int) {
	s.objs[i], s.objs[j] = s.objs[j], s.objs[i]
}
