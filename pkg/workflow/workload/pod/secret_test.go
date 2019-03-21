package pod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
)

type SecretSuite struct {
	suite.Suite
	client clientset.Interface
}

func (suite *SecretSuite) SetupSuite() {
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
}

func (suite *SecretSuite) TestParse() {
	assert := assert.New(suite.T())
	assert.Error(NewSecretRefValue().Parse(""))
	assert.Error(NewSecretRefValue().Parse("ns.secret"))
	assert.Error(NewSecretRefValue().Parse("$.secret"))
	assert.Error(NewSecretRefValue().Parse("$.ns.secret"))

	assert.Nil(NewSecretRefValue().Parse("$.ns.secret/a.b"))
	assert.Nil(NewSecretRefValue().Parse("$.ns.secret/a.b/c.d"))
}

func (suite *SecretSuite) TestResolve() {
	assert := assert.New(suite.T())
	refValue := NewSecretRefValue()
	assert.Nil(refValue.Parse("$.ns.secret/a.b"))
	v, err := refValue.Resolve(suite.client)
	assert.Nil(v)
	assert.Error(err)

	assert.Nil(refValue.Parse("$.ns.secret/data.key"))
	v, err = refValue.Resolve(suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	strV, ok := v.(string)
	assert.True(ok)
	assert.Equal("key1", strV)

	assert.Nil(refValue.Parse("$.ns.secret/data.json/user.languages[1]"))
	v, err = refValue.Resolve(suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	strV, ok = v.(string)
	assert.True(ok)
	assert.Equal("go", strV)

	assert.Nil(refValue.Parse("$.ns.secret/data.json/user.name"))
	v, err = refValue.Resolve(suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	strV, ok = v.(string)
	assert.True(ok)
	assert.Equal("cyclone", strV)
}

func TestSecretSuite(t *testing.T) {
	suite.Run(t, new(SecretSuite))
}
