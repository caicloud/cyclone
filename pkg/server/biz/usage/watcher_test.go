package usage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
)

type WatcherSuite struct {
	suite.Suite
	client clientset.Interface
	tenant string
}

func (suite *WatcherSuite) SetupTest() {
	suite.tenant = "cyclone"
	suite.client = fake.NewSimpleClientset()
}

func (suite *WatcherSuite) TestLaunchPVCUsageWatcher() {
	context := v1alpha1.ExecutionContext{
		Cluster:   "",
		Namespace: "test-namespace",
		PVC:       "test-pvc",
	}

	err := LaunchPVCUsageWatcher(suite.client, suite.tenant, context)
	assert.Nil(suite.T(), err)

	_, err = suite.client.AppsV1().Deployments(context.Namespace).Get(PVCWatcherName, metav1.GetOptions{})
	assert.Nil(suite.T(), err)
}

func (suite *WatcherSuite) TestDeletePVCUsageWatcher() {
	testNamespace := "test-namespace"
	_, err := suite.client.AppsV1().Deployments(testNamespace).Create(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: PVCWatcherName,
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	err = DeletePVCUsageWatcher(suite.client, testNamespace)
	assert.Nil(suite.T(), err)

	_, err = suite.client.AppsV1().Deployments(testNamespace).Get(PVCWatcherName, metav1.GetOptions{})
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not found")

}

func TestWatcherSuite(t *testing.T) {
	suite.Run(t, new(WatcherSuite))
}
