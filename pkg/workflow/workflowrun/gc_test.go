package workflowrun

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

func TestCheckGC(t *testing.T) {
	wfr := &v1alpha1.WorkflowRun{
		Status: v1alpha1.WorkflowRunStatus{
			Overall: v1alpha1.Status{
				Phase: v1alpha1.StatusRunning,
			},
			Cleaned: true,
		},
	}
	assert.False(t, checkGC(wfr))

	wfr = &v1alpha1.WorkflowRun{
		Status: v1alpha1.WorkflowRunStatus{
			Overall: v1alpha1.Status{
				Phase: v1alpha1.StatusRunning,
			},
			Cleaned: false,
		},
	}
	assert.False(t, checkGC(wfr))

	wfr = &v1alpha1.WorkflowRun{
		Status: v1alpha1.WorkflowRunStatus{
			Overall: v1alpha1.Status{
				Phase: v1alpha1.StatusCompleted,
			},
			Cleaned: false,
		},
	}
	assert.True(t, checkGC(wfr))

	wfr = &v1alpha1.WorkflowRun{
		Status: v1alpha1.WorkflowRunStatus{
			Overall: v1alpha1.Status{
				Phase: v1alpha1.StatusError,
			},
			Cleaned: false,
		},
	}
	assert.True(t, checkGC(wfr))
}

type GCProcessorSuite struct {
	suite.Suite
	processor *GCProcessor
}

func (s *GCProcessorSuite) SetupTest() {
	client := fake.NewSimpleClientset()
	recorder := new(MockedRecorder)
	recorder.On("Event", mock.Anything).Return()
	s.processor = &GCProcessor{
		client:   client,
		recorder: recorder,
		items:    make(map[string]*workflowRunItem),
		enabled:  true,
	}
}

func (s *GCProcessorSuite) TestAdd() {
	wfr := &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
		},
		Status: v1alpha1.WorkflowRunStatus{
			Overall: v1alpha1.Status{
				Phase: v1alpha1.StatusRunning,
			},
			Cleaned: true,
		},
	}
	s.processor.Add(wfr)
	assert.Nil(s.T(), s.processor.items["default:test1"])

	wfr = &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test2",
			Namespace: "default",
		},
		Status: v1alpha1.WorkflowRunStatus{
			Overall: v1alpha1.Status{
				Phase: v1alpha1.StatusCompleted,
			},
			Cleaned: false,
		},
	}
	s.processor.Add(wfr)
	assert.Equal(s.T(), "test2", s.processor.items["default:test2"].name)

	wfr = &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test3",
			Namespace: "default",
		},
		Status: v1alpha1.WorkflowRunStatus{
			Overall: v1alpha1.Status{
				Phase: v1alpha1.StatusError,
			},
			Cleaned: false,
		},
	}
	s.processor.Add(wfr)
	assert.Equal(s.T(), "test3", s.processor.items["default:test3"].name)
}

func (s *GCProcessorSuite) TestProcess() {
	pre := controller.Config.GC.DelaySeconds
	controller.Config.GC.DelaySeconds = 2
	defer func(preDelay time.Duration) {
		controller.Config.GC.DelaySeconds = pre
	}(pre)

	wfr := &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
		},
		Status: v1alpha1.WorkflowRunStatus{
			Overall: v1alpha1.Status{
				Phase:              v1alpha1.StatusCompleted,
				LastTransitionTime: metav1.Time{Time: time.Now()},
			},
			Cleaned: false,
		},
	}
	s.processor.Add(wfr)
	assert.Equal(s.T(), "test1", s.processor.items["default:test1"].name)
	s.processor.process()
	assert.Equal(s.T(), "test1", s.processor.items["default:test1"].name)
	time.Sleep(time.Second * 2)
	s.processor.process()
	assert.Nil(s.T(), s.processor.items["default:test1"])
}

func TestGCProcessorSuite(t *testing.T) {
	suite.Run(t, new(GCProcessorSuite))
}
