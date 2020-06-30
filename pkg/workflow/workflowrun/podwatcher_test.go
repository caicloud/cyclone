package workflowrun

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	k8stesting "k8s.io/client-go/testing"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
)

var (
	metaNs     = "meta"
	workloadNs = "workload"
	wfrName    = "test"
	stgName    = "stg"
	podName    = "pod"
	nowTime    = time.Now()
	timeBefore = metav1.Time{Time: nowTime.Add(100 * time.Second)}
	timeAfter  = metav1.Time{Time: nowTime.Add(200 * time.Second)}
)

type PodWatcherSuite struct {
	suite.Suite
	client clientset.Interface
}

// podWatcher implements watch.Interface
type podWatcher struct {
	c chan watch.Event
}

func newPodWatcher(fn func(cc chan watch.Event)) watch.Interface {
	p := &podWatcher{
		c: make(chan watch.Event),
	}

	go fn(p.c)
	return p
}

func (p *podWatcher) Stop() {
	close(p.c)
}

// Returns a chan which will receive all the events. If an error occurs
// or Stop() is called, this channel will be closed, in which case the
// watch should be completely cleaned up.
func (p *podWatcher) ResultChan() <-chan watch.Event {
	return p.c
}

func (suite *PodWatcherSuite) SetupTest() {
	client := fake.NewSimpleClientset()
	_, err := client.CycloneV1alpha1().WorkflowRuns(metaNs).Create(&v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wfrName,
			Namespace: metaNs,
		},
		Status: v1alpha1.WorkflowRunStatus{
			Stages: map[string]*v1alpha1.StageStatus{
				stgName: {},
			},
		},
	})
	assert.Nil(suite.T(), err)

	client.PrependWatchReactor("pods", func(action k8stesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, newPodWatcher(func(cc chan watch.Event) {
			e := watch.Event{
				Type: watch.Added,
				Object: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: workloadNs,
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			}

			cc <- e
			time.Sleep(100 * time.Millisecond)
			cc <- e
			time.Sleep(100 * time.Millisecond)
			e = watch.Event{
				Type: watch.Added,
				Object: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: workloadNs,
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				},
			}
			cc <- e
		}), nil
	})

	client.PrependWatchReactor("events", func(action k8stesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, newPodWatcher(func(cc chan watch.Event) {
			e := watch.Event{
				Type: watch.Added,
				Object: &corev1.Event{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-e-2",
						Namespace: workloadNs,
					},
					InvolvedObject: corev1.ObjectReference{
						Kind:      "pods",
						Namespace: workloadNs,
						Name:      podName,
					},
					Type:          "Warning",
					Reason:        "Failed",
					Message:       "Can not pull image xxx",
					LastTimestamp: timeBefore,
					Count:         1,
				},
			}
			cc <- e
			time.Sleep(10 * time.Millisecond)
			e = watch.Event{
				Type: watch.Added,
				Object: &corev1.Event{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-e-2",
						Namespace: workloadNs,
					},
					InvolvedObject: corev1.ObjectReference{
						Kind:      "pods",
						Namespace: workloadNs,
						Name:      podName,
					},
					Type:          "Warning",
					Reason:        "Failed",
					Message:       "Can not pull image xxx",
					LastTimestamp: timeBefore,
					Count:         2,
				},
			}
			cc <- e
			time.Sleep(10 * time.Millisecond)
			e = watch.Event{
				Type: watch.Added,
				Object: &corev1.Event{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-e-3",
						Namespace: workloadNs,
					},
					InvolvedObject: corev1.ObjectReference{
						Kind:      "pods",
						Namespace: workloadNs,
						Name:      podName,
					},
					Type:          "Warning",
					Reason:        "Failed",
					Message:       "Can not pull image xxx",
					LastTimestamp: timeAfter,
					Count:         1,
				},
			}
			cc <- e
		}), nil
	})

	suite.client = client
}

func (suite *PodWatcherSuite) TestPodEventWatch() {
	go newPodEventWatcher(suite.client, suite.client, metaNs, wfrName).Work(stgName, workloadNs, podName)
	// wait to work
	time.Sleep(1 * time.Second)
	wfr, err := suite.client.CycloneV1alpha1().WorkflowRuns(metaNs).Get(wfrName, metav1.GetOptions{})
	assert.Nil(suite.T(), err)
	events := wfr.Status.Stages[stgName].Events
	assert.NotEmpty(suite.T(), events)
	assert.Equal(suite.T(), 2, len(events))
	assert.Equal(suite.T(), "pod-e-3", events[0].Name)
	assert.Equal(suite.T(), int32(2), events[1].Count)
}

func TestPodWatcherSuite(t *testing.T) {
	suite.Run(t, new(PodWatcherSuite))
}
