package workflowrun

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
)

func TestKey(t *testing.T) {
	wfr := &v1alpha1.WorkflowRun{
		Spec: v1alpha1.WorkflowRunSpec{
			WorkflowRef: &corev1.ObjectReference{
				Namespace: "default",
				Name:      "test",
			},
		},
	}
	assert.Equal(t, "default/test", key(wfr))
}

type LimitQueuesSuite struct {
	suite.Suite
	queues *LimitedQueues
}

func (s *LimitQueuesSuite) SetupTest() {
	s.queues = NewLimitedQueues(fake.NewSimpleClientset(), 2)
}

func (s *LimitQueuesSuite) TestAddOrRefresh() {
	now := time.Now()
	wfr := &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "wfr1",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: now},
		},
		Spec: v1alpha1.WorkflowRunSpec{
			WorkflowRef: &corev1.ObjectReference{
				Namespace: "default",
				Name:      "wf1",
			},
		},
	}
	s.queues.AddOrRefresh(wfr)
	assert.Equal(s.T(), 1, s.queues.Queues[key(wfr)].size)
	refresh := s.queues.Queues[key(wfr)].head.next.refresh

	time.Sleep(time.Millisecond * 10)
	wfr = &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "wfr1",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: now},
		},
		Spec: v1alpha1.WorkflowRunSpec{
			WorkflowRef: &corev1.ObjectReference{
				Namespace: "default",
				Name:      "wf1",
			},
		},
	}
	s.queues.AddOrRefresh(wfr)
	assert.Equal(s.T(), 1, s.queues.Queues[key(wfr)].size)
	assert.NotEqual(s.T(), refresh, s.queues.Queues[key(wfr)].head.next.refresh)

	wfr = &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "wfr2",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: now.Add(time.Second)},
		},
		Spec: v1alpha1.WorkflowRunSpec{
			WorkflowRef: &corev1.ObjectReference{
				Namespace: "default",
				Name:      "wf1",
			},
		},
	}
	s.queues.AddOrRefresh(wfr)
	assert.Equal(s.T(), 2, s.queues.Queues[key(wfr)].size)

	wfr = &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "wfr3",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: now.Add(time.Second * 2)},
		},
		Spec: v1alpha1.WorkflowRunSpec{
			WorkflowRef: &corev1.ObjectReference{
				Namespace: "default",
				Name:      "wf1",
			},
		},
	}
	s.queues.AddOrRefresh(wfr)
	assert.Equal(s.T(), 2, s.queues.Queues[key(wfr)].size)
	assert.Equal(s.T(), now.Add(time.Second).Unix(), s.queues.Queues[key(wfr)].head.next.created)

	wfr = &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "wfr4",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: now.Add(-time.Second * 2)},
		},
		Spec: v1alpha1.WorkflowRunSpec{
			WorkflowRef: &corev1.ObjectReference{
				Namespace: "default",
				Name:      "wf1",
			},
		},
	}
	s.queues.AddOrRefresh(wfr)
	assert.Equal(s.T(), 2, s.queues.Queues[key(wfr)].size)
	assert.Equal(s.T(), now.Add(time.Second).Unix(), s.queues.Queues[key(wfr)].head.next.created)
}

func TestLimitQueuesSuite(t *testing.T) {
	suite.Run(t, new(LimitQueuesSuite))
}
