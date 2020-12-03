package cleaner

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	k8stesting "k8s.io/client-go/testing"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	serverv1alpha1 "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/util/k8s"
	"github.com/caicloud/cyclone/pkg/util/k8s/fake"
)

var (
	metaNs     = "meta"
	workloadNs = "workload"
	pName      = "test"
	nowTime    = time.Now()
	pvcName    = "test-pvc"
	startTime  = metav1.Time{Time: nowTime}
	endTime    = metav1.Time{Time: nowTime.Add(300 * time.Second)}
)

type CleanerSuite struct {
	suite.Suite
	client k8s.Interface
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

func (suite *CleanerSuite) SetupTest() {
	client := fake.NewSimpleClientset()

	client.PrependWatchReactor("pods", func(action k8stesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, newPodWatcher(func(cc chan watch.Event) {
			e := watch.Event{
				Type: watch.Added,
				Object: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      common.CacheCleanupContainerName,
						Namespace: workloadNs,
					},
					Status: corev1.PodStatus{
						Phase:     corev1.PodRunning,
						StartTime: &startTime,
					},
				},
			}

			cc <- e
			time.Sleep(100 * time.Millisecond)
			e = watch.Event{
				Type: watch.Modified,
				Object: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      common.CacheCleanupContainerName,
						Namespace: workloadNs,
					},
					Status: corev1.PodStatus{
						Phase:     corev1.PodSucceeded,
						StartTime: &startTime,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.CacheCleanupContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode:   0,
										FinishedAt: endTime,
									},
								},
							},
						},
					},
				},
			}
			cc <- e
		}), nil
	})

	suite.client = client
}

func (suite *CleanerSuite) TestClean() {
	client := suite.client
	_, err := client.CycloneV1alpha1().Projects(metaNs).Create(context.TODO(), &v1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pName,
			Namespace: metaNs,
		},
	}, metav1.CreateOptions{})
	assert.Nil(suite.T(), err)

	cleaner := NewCleaner(suite.client, suite.client, metaNs, pName)
	runningStatus, err := cleaner.Clean(workloadNs, pvcName)
	assert.Nil(suite.T(), err)

	assert.Equal(suite.T(), runningStatus.Phase, serverv1alpha1.CacheCleanupRunning)

	// wait to work
	time.Sleep(1 * time.Second)
	p, err := suite.client.CycloneV1alpha1().Projects(metaNs).Get(context.TODO(), pName, metav1.GetOptions{})
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), p.Annotations)
	assert.NotEmpty(suite.T(), p.Annotations[meta.AnnotationCacheCleanupStatus])

	var actual serverv1alpha1.CacheCleanupStatus
	err = json.Unmarshal([]byte(p.Annotations[meta.AnnotationCacheCleanupStatus]), &actual)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), actual.Acceleration.LatestStatus.Phase, serverv1alpha1.CacheCleanupSucceeded)
}

func (suite *CleanerSuite) TestInitCacheCleanupStatus() {
	client := suite.client
	runningCleanupStatus := serverv1alpha1.CacheCleanupStatus{
		Acceleration: serverv1alpha1.AccelerationCacheCleanupOverallStatus{
			LatestStatus: serverv1alpha1.AccelerationCacheCleanupStatus{
				TaskID:             "test",
				Phase:              serverv1alpha1.CacheCleanupRunning,
				StartTime:          startTime,
				LastTransitionTime: endTime,
			},
		},
	}

	ss, err := json.Marshal(runningCleanupStatus)
	assert.Nil(suite.T(), err)

	_, err = client.CycloneV1alpha1().Projects(metaNs).Create(context.TODO(), &v1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pName,
			Namespace: metaNs,
			Annotations: map[string]string{
				meta.AnnotationCacheCleanupStatus: string(ss),
			},
		},
	}, metav1.CreateOptions{})
	assert.Nil(suite.T(), err)

	podName := generatePodName(pName)
	_, err = client.CoreV1().Pods(metaNs).Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: metaNs,
			Labels: map[string]string{
				meta.LabelPodKind:      meta.PodKindAccelerationGC.String(),
				meta.LabelPodCreatedBy: meta.CycloneCreator,
			},
			Annotations: map[string]string{
				meta.AnnotationIstioInject: meta.AnnotationValueFalse,
			},
		},
	}, metav1.CreateOptions{})
	assert.Nil(suite.T(), err)

	_, err = client.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: metaNs,
			Labels: map[string]string{
				meta.LabelTenantName: "t1",
			},
		},
	}, metav1.CreateOptions{})
	assert.Nil(suite.T(), err)

	err = InitCacheCleanupStatus(suite.client)
	assert.Nil(suite.T(), err)
	p, err := suite.client.CycloneV1alpha1().Projects(metaNs).Get(context.TODO(), pName, metav1.GetOptions{})
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), p.Annotations)
	initStatusString, ok := p.Annotations[meta.AnnotationCacheCleanupStatus]
	assert.True(suite.T(), ok)

	var initStatus serverv1alpha1.CacheCleanupStatus
	err = json.Unmarshal([]byte(initStatusString), &initStatus)
	assert.Nil(suite.T(), err)
	assert.EqualValues(suite.T(), serverv1alpha1.CacheCleanupFailed, initStatus.Acceleration.LatestStatus.Phase)

	_, err = suite.client.CoreV1().Pods(metaNs).Get(context.TODO(), podName, metav1.GetOptions{})
	assert.NotNil(suite.T(), err)
	assert.True(suite.T(), k8serr.IsNotFound(err))
}

func (suite *CleanerSuite) TestStopReasonNoNeed() {
	client := suite.client
	podName := "pod1"
	_, err := client.CoreV1().Pods(metaNs).Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
			Labels: map[string]string{
				meta.LabelPodKind: meta.PodKindAccelerationGC.String(),
			},
		},
	}, metav1.CreateOptions{})
	assert.Nil(suite.T(), err)

	err = StopReasonNoNeed(client, metaNs, NoNeedReasonPVCDeleted)
	assert.Nil(suite.T(), err)

	_, err = client.CoreV1().Pods(metaNs).Get(context.TODO(), podName, metav1.GetOptions{})
	assert.NotNil(suite.T(), err)
	assert.True(suite.T(), k8serr.IsNotFound(err))
}

func TestCleanerSuite(t *testing.T) {
	suite.Run(t, new(CleanerSuite))
}
