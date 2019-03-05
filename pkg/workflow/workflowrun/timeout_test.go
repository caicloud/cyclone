package workflowrun

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
)

func TestParseTime(t *testing.T) {
	cases := []struct {
		input    string
		expected time.Duration
		err      bool
	}{
		{
			input:    "30min",
			expected: time.Minute * 30,
		},
		{
			input:    "20m",
			expected: time.Minute * 20,
		},
		{
			input:    "30s",
			expected: time.Second * 30,
		},
		{
			input:    "1hour",
			expected: time.Hour,
		},
		{
			input:    "1h30s",
			expected: time.Hour + time.Second*30,
		},
		{
			input:    "1h30m",
			expected: time.Hour + time.Minute*30,
		},
		{
			input:    "1h30m30s",
			expected: time.Hour + time.Minute*30 + time.Second*30,
		},
		{
			input: "1h:30m:30s",
			err:   true,
		},
		{
			input:    "1H30m",
			expected: time.Hour + time.Minute*30,
		},
		{
			input:    "1H30MIN",
			expected: time.Hour + time.Minute*30,
		},
		{
			input: "1H30MIN----",
			err:   true,
		},
		{
			input: "1millisecond",
			err:   true,
		},
		{
			input: "",
			err:   true,
		},
	}

	for _, c := range cases {
		out, err := ParseTime(c.input)
		if c.err {
			if err == nil {
				t.Errorf("%s expected to be invalid, but get %v", c.input, out)
			}
		} else {
			if out != c.expected {
				t.Errorf("%s expected to be %v, but got %v", c.input, c.expected, out)
			}
		}
	}
}

func TestNewWorkflowRunItem(t *testing.T) {
	wfr := &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.WorkflowRunSpec{
			Timeout: "30s",
		},
	}
	result := newWorkflowRunItem(wfr)
	expected := &workflowRunItem{
		name:       "test",
		namespace:  "default",
		expireTime: time.Now().Add(time.Second * 30),
	}

	if expected.name != result.name ||
		expected.namespace != result.namespace ||
		(expected.expireTime.Unix()-result.expireTime.Unix()) > 1 {
		t.Errorf("%v expected, but got %v", expected, result)
	}
}

type MockedRecorder struct {
	mock.Mock
}

func (r *MockedRecorder) Event(object runtime.Object, eventtype, reason, message string) {
}
func (r *MockedRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}
func (r *MockedRecorder) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}) {
}
func (r *MockedRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}

type TimeoutProcessorSuite struct {
	suite.Suite
	processor *TimeoutProcessor
}

func (suite *TimeoutProcessorSuite) SetupTest() {
	client := fake.NewSimpleClientset()
	recorder := new(MockedRecorder)
	recorder.On("Event", mock.Anything).Return()
	suite.processor = &TimeoutProcessor{
		client:   client,
		recorder: recorder,
		items:    make(map[string]*workflowRunItem),
	}
}

func (suite *TimeoutProcessorSuite) TestAdd() {
	suite.processor.Add(&v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
		},
		Spec: v1alpha1.WorkflowRunSpec{
			Timeout: "30s",
		},
	})
	assert.Equal(suite.T(), 1, len(suite.processor.items))
	assert.Equal(suite.T(), "test1", suite.processor.items["default:test1"].name)

	suite.processor.Add(&v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test2",
			Namespace: "default",
		},
		Spec: v1alpha1.WorkflowRunSpec{
			Timeout: "30s",
		},
	})

	assert.Equal(suite.T(), 2, len(suite.processor.items))
	assert.Equal(suite.T(), "test1", suite.processor.items["default:test1"].name)
	assert.Equal(suite.T(), "test2", suite.processor.items["default:test2"].name)
}

func (suite *TimeoutProcessorSuite) TestProcess() {
	suite.processor.Add(&v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
		},
		Spec: v1alpha1.WorkflowRunSpec{
			Timeout: "1s",
		},
		Status: v1alpha1.WorkflowRunStatus{
			Stages: map[string]*v1alpha1.StageStatus{
				"stg1": {},
				"stg2": {
					Pod: &v1alpha1.PodInfo{
						Name:      "stg1",
						Namespace: "default",
					},
				},
			},
		},
	})
	suite.processor.Add(&v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test2",
			Namespace: "default",
		},
		Spec: v1alpha1.WorkflowRunSpec{
			Timeout: "30s",
		},
	})
	suite.Equal(2, len(suite.processor.items))

	time.Sleep(time.Second)
	suite.processor.process()
	suite.Equal(1, len(suite.processor.items))
	suite.Nil(suite.processor.items["default:test1"])
}

func TestTimeoutProcessorSuite(t *testing.T) {
	suite.Run(t, new(TimeoutProcessorSuite))
}
