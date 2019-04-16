package sorter

import (
	corev1 "k8s.io/api/core/v1"
)

// NamespaceSort implements the sort.Interface{}
type NamespaceSort struct {
	objs      []corev1.Namespace
	ascending bool
}

// NewNamespaceSorter ...
func NewNamespaceSorter(objs []corev1.Namespace, ascending bool) *NamespaceSort {
	return &NamespaceSort{
		objs:      objs,
		ascending: ascending,
	}
}

// Len is the number of elements in the collection.
func (s *NamespaceSort) Len() int {
	return len(s.objs)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s *NamespaceSort) Less(i, j int) bool {
	if s.ascending {
		return s.objs[i].CreationTimestamp.Before(&s.objs[j].CreationTimestamp)
	}
	return s.objs[i].CreationTimestamp.After(s.objs[j].CreationTimestamp.Time)
}

// Swap swaps the elements with indexes i and j.
func (s *NamespaceSort) Swap(i, j int) {
	s.objs[i], s.objs[j] = s.objs[j], s.objs[i]
}
