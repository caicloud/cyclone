package ref

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

var (
	secretRegexpString = `^\${secrets.([-_\w\.]+):([-_\w\.]+)/(\S+)}$`
	secretRegexp       = regexp.MustCompile(secretRegexpString)
)

const (
	secretTypeRefFormat = `${secrets.<namespace>:<secret-name>/<jsonpath>/<jsonpath>/...}`
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
// ${secrets.<namespace>:<secret-name>/<jsonpath>/<jsonpath>}
// For example, in secret (named 'secret' under namespace 'ns'):
// {
//   "apiVersion": "v1",
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
// ${secrets.ns:secret/data.key}  --> KEY
// ${secrets.ns:secret/data.json/user.id}  --> 111
func (r *SecretRefValue) Parse(ref string) error {
	if err := r.parseNew(ref); err != nil {
		// for compatibility, will remove in the future
		if err = r.parseOld(ref); err != nil {
			return err
		}
	}

	return nil
}

func (r *SecretRefValue) parseNew(ref string) error {
	trimed := strings.TrimSpace(ref)
	results := secretRegexp.FindStringSubmatch(trimed)
	if len(results) < 4 {
		return fmt.Errorf("secret type ref must be specified as %s, but got '%s'", secretTypeRefFormat, ref)
	}

	r.Namespace = results[1]
	r.Secret = results[2]

	r.Jsonpaths = strings.Split(results[3], "/")
	return nil
}

// ParseOld parses a given ref. The reference value specifies json path
// in a secret. Format of the reference is:
// $.<namespace>.<secret-name>/<jsonpath>/<jsonpath>
// For example, in secret (named 'secret' under namespace 'ns'):
// {
//   "apiVersion": "v1",
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
//
// +Deprecated, maybe removed in the future.
func (r *SecretRefValue) parseOld(ref string) error {
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
func (r *SecretRefValue) Resolve(client clientset.Interface) (string, error) {
	if len(r.Secret) == 0 || len(r.Jsonpaths) == 0 {
		return "", errors.New("empty secret name or jsonpath")
	}

	if len(r.Namespace) == 0 {
		r.Namespace = "default"
	}

	s, err := client.CoreV1().Secrets(r.Namespace).Get(r.Secret, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get secret %s:%s error: %v", r.Namespace, r.Secret, err)
	}

	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	var v interface{}
	for index, path := range r.Jsonpaths {
		var obj interface{}
		err := json.Unmarshal(data, &obj)
		if err != nil {
			return "", err
		}
		v, err = jsonpath.Get(path, obj)
		if err != nil {
			return "", err
		}

		strV, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("expect string value, got: %v", v)
		}
		if index == 0 {
			data, err = base64.StdEncoding.DecodeString(strV)
			if err != nil {
				return "", err
			}
			v = string(data)
		} else {
			data = []byte(strV)
		}
	}

	strV, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("expect string value, got: %v", v)
	}
	return strV, nil
}
