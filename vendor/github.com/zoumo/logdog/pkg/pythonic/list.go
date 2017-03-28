package pythonic

import "bytes"

// List is a python-like type
type List []interface{}

// NewList returns a new list
func NewList(capacity int) List {
	return make(List, 0, capacity)
}

// Append to the end of list
func (l *List) Append(item ...interface{}) List {
	*l = append(*l, item...)
	return *l
}

// Extend list with another list
func (l *List) Extend(list List) List {
	*l = append(*l, list...)
	return *l
}

// String convert list to a string
func (l List) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("[")
	for i, v := range l {
		buf.WriteString(spprint(v))
		if i < len(l)-1 {
			buf.WriteString(", ")

		}
	}
	buf.WriteString("]")
	return buf.String()

}
