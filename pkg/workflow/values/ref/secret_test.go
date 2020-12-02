package ref

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"github.com/caicloud/cyclone/pkg/util/k8s"
	"github.com/caicloud/cyclone/pkg/util/k8s/fake"
)

type SecretSuite struct {
	suite.Suite
	client k8s.Interface
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
	secretRefValue := NewSecretRefValue()
	assert.Error(secretRefValue.Parse(""))
	assert.Error(secretRefValue.Parse("ns.secret"))
	assert.Error(secretRefValue.Parse("${secret}"))
	assert.Error(secretRefValue.Parse("${ns.secret}"))

	assert.Nil(secretRefValue.Parse("${secrets.ns:secret/a.b}"))
	assert.Nil(secretRefValue.Parse("${secrets.ns:secret/a.b/c.d}"))
}

func (suite *SecretSuite) TestParseOld() {
	assert := assert.New(suite.T())
	secretRefValue := NewSecretRefValue()
	assert.Error(secretRefValue.Parse(""))
	assert.Error(secretRefValue.Parse("ns.secret"))
	assert.Error(secretRefValue.Parse("$.secret"))
	assert.Error(secretRefValue.Parse("$.ns.secret"))

	assert.Nil(secretRefValue.Parse("$.ns.secret/a.b"))
	assert.Nil(secretRefValue.Parse("$.ns.secret/a.b/c.d"))
}

func (suite *SecretSuite) TestResolveOld() {
	assert := assert.New(suite.T())
	secretRefValue := NewSecretRefValue()
	assert.Nil(secretRefValue.Parse("$.ns.secret/a.b"))
	v, err := secretRefValue.Resolve(suite.client)
	assert.Empty(v)
	assert.Error(err)

	assert.Nil(secretRefValue.Parse("$.ns.secret/data.key"))
	v, err = secretRefValue.Resolve(suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("key1", v)

	assert.Nil(secretRefValue.Parse("$.ns.secret/data.json/user.languages[1]"))
	v, err = secretRefValue.Resolve(suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("go", v)

	assert.Nil(secretRefValue.Parse("$.ns.secret/data.json/user.name"))
	v, err = secretRefValue.Resolve(suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("cyclone", v)
}

func (suite *SecretSuite) TestResolve() {
	assert := assert.New(suite.T())
	secretRefValue := NewSecretRefValue()
	assert.Nil(secretRefValue.Parse("${secrets.ns:secret/a.b}"))
	v, err := secretRefValue.Resolve(suite.client)
	assert.Empty(v)
	assert.Error(err)

	assert.Nil(secretRefValue.Parse("${secrets.ns:secret/data.key}"))
	v, err = secretRefValue.Resolve(suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("key1", v)

	assert.Nil(secretRefValue.Parse("${secrets.ns:secret/data.json/user.languages[1]}"))
	v, err = secretRefValue.Resolve(suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("go", v)

	assert.Nil(secretRefValue.Parse("${secrets.ns:secret/data.json/user.name}"))
	v, err = secretRefValue.Resolve(suite.client)
	assert.NotNil(v)
	assert.Nil(err)
	assert.Equal("cyclone", v)
}

func TestSecretSuite(t *testing.T) {
	suite.Run(t, new(SecretSuite))
}
