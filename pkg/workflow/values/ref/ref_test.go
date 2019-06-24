package ref

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
)

type RefSuite struct {
	suite.Suite
	wfr    *v1alpha1.WorkflowRun
	client clientset.Interface
}

func (suite *RefSuite) SetupSuite() {
	jsonData := "{\"user\": {\"name\":\"cyclone\", \"languages\": [\"java\", \"go\"]}}"
	client := fake.NewSimpleClientset()
	client.PrependReactor("get", "secrets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "ns",
			},
			Data: map[string][]byte{
				"key":  []byte("key1"),
				"json": []byte(jsonData),
			},
		}, nil
	})
	suite.client = client

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

func (suite *RefSuite) TestResolveRefStringValue() {
	assert := assert.New(suite.T())
	processor := NewProcess(suite.wfr)

	v, err := processor.ResolveRefStringValue("", suite.client)
	assert.Nil(err)
	assert.Equal("", v)

	v, err = processor.ResolveRefStringValue("test1", suite.client)
	assert.Nil(err)
	assert.Equal("test1", v)

	v, err = processor.ResolveRefStringValue("${variables.Registry}", suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("docker.io", v)

	v, err = processor.ResolveRefStringValue("${variables.IMAGE}", suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("cyclone", v)

	v, err = processor.ResolveRefStringValue("${variables.Registry}", suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("docker.io", v)

	v, err = processor.ResolveRefStringValue("${secrets.ns:secret/a.b}", suite.client)
	assert.NotNil(v)
	assert.Error(err)

	v, err = processor.ResolveRefStringValue("${secrets.ns:secret/data.key}", suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("key1", v)

	v, err = processor.ResolveRefStringValue("${secrets.ns:secret/data.json/user.languages[1]}", suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("go", v)

	v, err = processor.ResolveRefStringValue("${secrets.ns:secret/data.json/user.name}", suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("cyclone", v)
}

func TestRefSuite(t *testing.T) {
	suite.Run(t, new(RefSuite))
}
