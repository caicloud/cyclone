package ref

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

type VariableSuite struct {
	suite.Suite
	wfr *v1alpha1.WorkflowRun
}

func (suite *VariableSuite) SetupSuite() {
	suite.wfr = &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-wfr",
		},
		Spec: v1alpha1.WorkflowRunSpec{
			GlobalVariables: []v1alpha1.GlobalVariable{{
				Name:  "Registry",
				Value: "docker.io",
			},
				{
					Name:  "IMAGE",
					Value: "cyclone",
				},
			},
		},
	}
}

func (suite *VariableSuite) TestParse() {
	assert := assert.New(suite.T())
	refValue := NewVariableRefValue(suite.wfr)
	assert.Error(refValue.Parse(""))
	assert.Error(refValue.Parse("variables."))
	assert.Error(refValue.Parse("${variables}"))
	assert.Error(refValue.Parse("${variables.secret/}"))

	assert.Nil(refValue.Parse("${variables.Registry}"))
	assert.Nil(refValue.Parse("${variables.IMAGE}"))
}

func (suite *VariableSuite) TestResolve() {
	assert := assert.New(suite.T())
	refValue := NewVariableRefValue(suite.wfr)
	assert.Nil(refValue.Parse("${variables.test}"))
	v, err := refValue.Resolve()
	assert.Empty(v)
	assert.Error(err)

	assert.Nil(refValue.Parse("${variables.Registry}"))
	v, err = refValue.Resolve()
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("docker.io", v)

	assert.Nil(refValue.Parse("${variables.IMAGE}"))
	v, err = refValue.Resolve()
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("cyclone", v)

	refValueNil := NewVariableRefValue(nil)
	assert.Nil(refValueNil.Parse("${variables.IMAGE}"))
	v, err = refValueNil.Resolve()
	assert.Empty(v)
	assert.Error(err)
}

func TestVariableSuite(t *testing.T) {
	suite.Run(t, new(VariableSuite))
}
