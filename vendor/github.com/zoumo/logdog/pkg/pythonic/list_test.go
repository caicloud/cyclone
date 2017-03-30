package pythonic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	list := NewList(1)
	assert.Equal(t, 1, cap(list))
	list2 := NewList(2)
	assert.Equal(t, 2, cap(list2))
	list.Append(1, 3, 4, 5, 6)
	assert.Equal(t, List{1, 3, 4, 5, 6}, list)
	list2.Append("2342", 423.546, "xxx")
	list = list.Extend(list2)
	assert.Equal(t, List{1, 3, 4, 5, 6, "2342", 423.546, "xxx"}, list)
}
