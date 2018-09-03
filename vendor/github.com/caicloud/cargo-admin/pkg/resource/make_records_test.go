package resource

import (
	"reflect"
	"testing"
	"time"

	"github.com/caicloud/cargo-admin/pkg/harbor"

	"github.com/davecgh/go-spew/spew"
)

func TestSplitRecords(t *testing.T) {
	now := time.Now()
	cases := []struct {
		jobs     []*harbor.HarborRepJob
		expected [][]*harbor.HarborRepJob
	}{
		{
			jobs: []*harbor.HarborRepJob{
				&harbor.HarborRepJob{CreationTime: now},
			},
			expected: [][]*harbor.HarborRepJob{
				[]*harbor.HarborRepJob{
					&harbor.HarborRepJob{CreationTime: now},
				},
			},
		},
		{
			jobs: []*harbor.HarborRepJob{
				&harbor.HarborRepJob{CreationTime: now},
				&harbor.HarborRepJob{CreationTime: now.Add(time.Second * 1)},
				&harbor.HarborRepJob{CreationTime: now.Add(time.Second * 2)},
			},
			expected: [][]*harbor.HarborRepJob{
				[]*harbor.HarborRepJob{
					&harbor.HarborRepJob{CreationTime: now},
				},
				[]*harbor.HarborRepJob{
					&harbor.HarborRepJob{CreationTime: now.Add(time.Second * 1)},
				},
				[]*harbor.HarborRepJob{
					&harbor.HarborRepJob{CreationTime: now.Add(time.Second * 2)},
				},
			},
		},
		{
			jobs: []*harbor.HarborRepJob{
				&harbor.HarborRepJob{CreationTime: now},
				&harbor.HarborRepJob{CreationTime: now.Add(time.Millisecond * 500)},
				&harbor.HarborRepJob{CreationTime: now.Add(time.Second * 1)},
			},
			expected: [][]*harbor.HarborRepJob{
				[]*harbor.HarborRepJob{
					&harbor.HarborRepJob{CreationTime: now},
					&harbor.HarborRepJob{CreationTime: now.Add(time.Millisecond * 500)},
					&harbor.HarborRepJob{CreationTime: now.Add(time.Second * 1)},
				},
			},
		},
		{
			jobs: []*harbor.HarborRepJob{
				&harbor.HarborRepJob{CreationTime: now},
				&harbor.HarborRepJob{CreationTime: now.Add(time.Millisecond * 500)},
				&harbor.HarborRepJob{CreationTime: now.Add(time.Second * 2)},
			},
			expected: [][]*harbor.HarborRepJob{
				[]*harbor.HarborRepJob{
					&harbor.HarborRepJob{CreationTime: now},
					&harbor.HarborRepJob{CreationTime: now.Add(time.Millisecond * 500)},
				},
				[]*harbor.HarborRepJob{
					&harbor.HarborRepJob{CreationTime: now.Add(time.Second * 2)},
				},
			},
		},
	}

	for _, c := range cases {
		result := splitRepJobs(c.jobs)
		if !reflect.DeepEqual(result, c.expected) {
			t.Errorf("Split jobs error, expected %s, but got %s", spew.Sdump(c.expected), spew.Sdump(result))
		}
	}
}
