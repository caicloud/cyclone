package usage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
)

type WatcherSuite struct {
	suite.Suite
	client clientset.Interface
	tenant string
}

func (suite *WatcherSuite) SetupTest() {
	suite.tenant = "cyclone"
	client := fake.NewSimpleClientset()
	Init(client)

	_, err := client.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: common.TenantNamespace(suite.tenant),
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	suite.client = client
}

func (suite *WatcherSuite) TestLaunchPVCUsageWatcher() {
	watcherCtl := NewWatcherController(suite.client, suite.tenant)
	context := v1alpha1.ExecutionContext{
		Cluster:   "",
		Namespace: "test-namespace",
		PVC:       "test-pvc",
	}

	err := watcherCtl.LaunchPVCUsageWatcher(context)
	assert.Nil(suite.T(), err)

	_, err = suite.client.ExtensionsV1beta1().Deployments(context.Namespace).Get(PVCWatcherName, metav1.GetOptions{})
	assert.Nil(suite.T(), err)

	actualResource, err := watcherCtl.recorder.GetWatcherResource()
	assert.Nil(suite.T(), err)

	watcherConfig := config.Config.StorageUsageWatcher
	expectResource := &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(getOrDefault(&watcherConfig, corev1.ResourceRequestsCPU, "50m")),
			corev1.ResourceMemory: resource.MustParse(getOrDefault(&watcherConfig, corev1.ResourceRequestsMemory, "32Mi")),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(getOrDefault(&watcherConfig, corev1.ResourceLimitsCPU, "100m")),
			corev1.ResourceMemory: resource.MustParse(getOrDefault(&watcherConfig, corev1.ResourceLimitsMemory, "128Mi")),
		},
	}
	assert.Equal(suite.T(), expectResource, actualResource)
}

func (suite *WatcherSuite) TestDeletePVCUsageWatcher() {
	testNamespace := "test-namespace"
	_, err := suite.client.ExtensionsV1beta1().Deployments(testNamespace).Create(&v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: PVCWatcherName,
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	watcherCtl := NewWatcherController(suite.client, suite.tenant)
	err = watcherCtl.DeletePVCUsageWatcher(testNamespace)
	assert.Nil(suite.T(), err)

	_, err = suite.client.ExtensionsV1beta1().Deployments(testNamespace).Get(PVCWatcherName, metav1.GetOptions{})
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not found")

	actualResource, err := watcherCtl.recorder.GetWatcherResource()
	assert.Nil(suite.T(), err)

	zeroQuantity := resource.MustParse("0")
	expectResource := &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    zeroQuantity,
			corev1.ResourceMemory: zeroQuantity,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    zeroQuantity,
			corev1.ResourceMemory: zeroQuantity,
		},
	}

	assert.Equal(suite.T(), expectResource, actualResource)
}

func TestWatcherSuite(t *testing.T) {
	suite.Run(t, new(WatcherSuite))
}
