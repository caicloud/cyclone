package usage

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/util/k8s"
	"github.com/caicloud/cyclone/pkg/util/k8s/fake"
)

type WatcherSuite struct {
	suite.Suite
	client k8s.Interface
	tenant string
}

func (suite *WatcherSuite) SetupTest() {
	suite.tenant = "cyclone"
	suite.client = fake.NewSimpleClientset()
}

func (suite *WatcherSuite) TestLaunchPVCUsageWatcher() {
	execCtx := v1alpha1.ExecutionContext{
		Cluster:   "",
		Namespace: "test-namespace",
		PVC:       "test-pvc",
	}

	err := LaunchPVCUsageWatcher(suite.client, suite.tenant, execCtx)
	assert.Nil(suite.T(), err)

	_, err = suite.client.AppsV1().Deployments(execCtx.Namespace).Get(context.TODO(), PVCWatcherName, metav1.GetOptions{})
	assert.Nil(suite.T(), err)
}

func (suite *WatcherSuite) TestDeletePVCUsageWatcher() {
	testNamespace := "test-namespace"
	_, err := suite.client.AppsV1().Deployments(testNamespace).Create(context.TODO(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: PVCWatcherName,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
	}

	err = DeletePVCUsageWatcher(suite.client, testNamespace)
	assert.Nil(suite.T(), err)

	_, err = suite.client.AppsV1().Deployments(testNamespace).Get(context.TODO(), PVCWatcherName, metav1.GetOptions{})
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not found")

}

func TestWatcherSuite(t *testing.T) {
	suite.Run(t, new(WatcherSuite))
}
