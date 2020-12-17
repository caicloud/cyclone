package ref

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

// SecretGetter is used to get a secret from the cluster.
// The reason we use SecretGetter here is to decouple from the kubernetes clientset.
type SecretGetter = func(namespace, name string) (*corev1.Secret, error)

// Processor processes ref value to a string
type Processor struct {
	wfr              *v1alpha1.WorkflowRun
	secretRefValue   *SecretRefValue
	variableRefValue *VariableRefValue
	secretGetter     SecretGetter
}

// NewProcessor creates a processor object
func NewProcessor(wfr *v1alpha1.WorkflowRun, getter SecretGetter) *Processor {
	return &Processor{
		wfr:              wfr,
		secretRefValue:   NewSecretRefValue(),
		variableRefValue: NewVariableRefValue(wfr),
		secretGetter:     getter,
	}
}

// ResolveRefStringValue resolves a ref kind value, if it's not a ref value, return the origin value.
// Valid ref values supported now include:
// - '${secrets.<ns>:<secret>/<jsonpath>/...}' to refer value in a secret
// - '${stages.<stage>.outputs.<key>}' to refer value from a stage output
// - '${variables.<key>}' to refer value from a global variable defined in wfr
func (p *Processor) ResolveRefStringValue(ref string) (string, error) {
	var value string
	var err error

	if err = p.secretRefValue.Parse(ref); err == nil {
		value, err = p.secretRefValue.Resolve(p.secretGetter)
		if err != nil {
			return ref, err
		}
	} else if err = p.variableRefValue.Parse(ref); err == nil {
		value, err = p.variableRefValue.Resolve()
		if err != nil {
			return ref, err
		}
	} else {
		// TODO(ChenDe): implement ${stages.<stage>.outputs.<key>} type ref value
		return ref, nil
	}

	return value, nil
}
