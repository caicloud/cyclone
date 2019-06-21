package ref

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

var (
	variableRegexpString = `^\${variables.([-_\w\.]+)}$`
	variableRegexp       = regexp.MustCompile(variableRegexpString)
)

const (
	variableTypeRefFormat = `${variables.<key>}`
)

// VariableRefValue represents a global(wfr scope) variable value. It's defined in wfr spec.globalVariables.
type VariableRefValue struct {
	wfr *v1alpha1.WorkflowRun
	// Name of the variable
	Name string
}

// NewVariableRefValue create a variable reference value.
func NewVariableRefValue(wfr *v1alpha1.WorkflowRun) *VariableRefValue {
	return &VariableRefValue{
		wfr: wfr,
	}
}

// Parse parses a given ref. The reference value specifies variable key defined in wfr. Format of the reference is:
// ${variables.<key>}
//
// For example, in wfr (named 'secret' under namespace 'ns'):
// {
//   "kind": "WorkflowRun",
//   ...
//   "spec": {
//     "globalVariables": [
//       {
//         "IMAGE": "cyclone"
//       },
//       {
//         "Registry": "docker.io"
//       }
//     ]
//   }
// }
// ${variables.IMAGE}  --> cyclone
// ${variables.Registry}  --> docker.io
func (r *VariableRefValue) Parse(ref string) error {
	trimed := strings.TrimSpace(ref)
	results := variableRegexp.FindStringSubmatch(trimed)
	if len(results) < 2 {
		return fmt.Errorf("variable type ref must be specified as %s, but got '%s'", variableTypeRefFormat, ref)
	}

	r.Name = results[1]
	return nil
}

// Resolve resolves the variable ref and get the real value.
func (r *VariableRefValue) Resolve() (string, error) {
	if r.wfr == nil {
		return "", fmt.Errorf("wfr is nil")
	}

	for _, variable := range r.wfr.Spec.GlobalVariables {
		if variable.Name == r.Name {
			return variable.Value, nil
		}
	}

	return "", fmt.Errorf("not found variable %s in wfr", r.Name)
}
