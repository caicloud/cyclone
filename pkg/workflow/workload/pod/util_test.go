package pod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
)

func TestGetResourceVolumeName(t *testing.T) {
	assert.Equal(t, "rsc-git", GetResourceVolumeName("git"))
}

func TestResolveRefStringValue(t *testing.T) {
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

	assert := assert.New(t)
	v, err := ResolveRefStringValue("value1", client)
	assert.Nil(err)
	assert.Equal("value1", v)
	v, err = ResolveRefStringValue("111", client)
	assert.Nil(err)
	assert.Equal("111", v)
	v, err = ResolveRefStringValue("ns.secret", client)
	assert.Nil(err)
	assert.Equal("ns.secret", v)
	v, err = ResolveRefStringValue("$.ns.secret/data.key", client)
	assert.Nil(err)
	assert.Equal("key1", v)
	_, err = ResolveRefStringValue("$.ns.secret/data.non-exist-key", client)
	assert.Error(err)
	_, err = ResolveRefStringValue("$.ns.secret/data.json/non-exist-key", client)
	assert.Error(err)
	v, err = ResolveRefStringValue("$.ns.secret/data.json/user.name", client)
	assert.Nil(err)
	assert.Equal("cyclone", v)
}
