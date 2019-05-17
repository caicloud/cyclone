package pod

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

// SecretRefValue represents a value in a secret. It's defined by a secret and json paths.
type SecretRefValue struct {
	// Namespace of the secret
	Namespace string
	// Name of the secret
	Secret string
	// Json path in the secret to refer the value. If more than one
	// paths provided, value resolved for the previous path would be
	// regarded as a marshaled json and be used as data source to the
	// following one.
	Jsonpaths []string
}

// NewSecretRefValue create a secret reference value.
func NewSecretRefValue() *SecretRefValue {
	return &SecretRefValue{}
}

// Parse parses a given ref. The reference value specifies json path
// in a secret. Format of the reference is:
// $.<namespace>.<secret-name>/<jsonpath>/<jsonpath>
// For example, in secret (named 'secret' under namespace 'ns'):
// {
//  "apiVersion": "v1",
//   "data": {
//    "key": "KEY",
//    "json": "{\"user\":{\"id\": \"111\"}}"
//   },
//   "kind": "Secret",
//   "type": "Opaque",
//   "metadata": {
//       ...
//   }
// }
// $.ns.secret/data.key  --> KEY
// $.ns.secret/data.json/user.id  --> 111
func (r *SecretRefValue) Parse(ref string) error {
	if !strings.HasPrefix(ref, "$.") {
		return fmt.Errorf("ref should begin with '$.': %s", ref)
	}

	parts := strings.Split(strings.Trim(ref, " "), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid ref: %s", ref)
	}

	segs := strings.Split(parts[0], ".")
	if len(segs) != 3 {
		return fmt.Errorf("secret must be specified as '$.<ns>.<secret>', but got '%s'", parts[0])
	}
	r.Namespace = segs[1]
	r.Secret = segs[2]

	r.Jsonpaths = parts[1:]
	return nil
}

// Resolve resolves the secret ref and get the real value.
func (r *SecretRefValue) Resolve(client clientset.Interface) (interface{}, error) {
	if len(r.Secret) == 0 || len(r.Jsonpaths) == 0 {
		return nil, errors.New("empty secret name or jsonpath")
	}

	if len(r.Namespace) == 0 {
		r.Namespace = "default"
	}

	s, err := client.CoreV1().Secrets(r.Namespace).Get(r.Secret, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get secret %s:%s error: %v", r.Namespace, r.Secret, err)
	}

	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var v interface{}
	for index, path := range r.Jsonpaths {
		var obj interface{}
		err := json.Unmarshal(data, &obj)
		if err != nil {
			return nil, err
		}
		v, err = jsonpath.Get(path, obj)
		if err != nil {
			return nil, err
		}

		strV, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expect string value, got: %v", v)
		}
		if index == 0 {
			data, err = base64.StdEncoding.DecodeString(strV)
			if err != nil {
				return nil, err
			}
			v = string(data)
		} else {
			data = []byte(strV)
		}
	}

	return v, nil
}
