package delegation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

// Request is request sent to delegation service.
type Request struct {
	Stage       *v1alpha1.Stage       `json:"stage"`
	Workflow    *v1alpha1.Workflow    `json:"workflow"`
	WorkflowRun *v1alpha1.WorkflowRun `json:"workflowrun"`
}

// Delegate sends request to delegation service.
func Delegate(request *Request) error {
	delegation := request.Stage.Spec.Delegation
	log.WithField("stg", request.Stage.Name).Info("Delegate stage to: ", delegation.URL)

	raw, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request error: %v", err)
	}

	rsp, err := http.DefaultClient.Post(delegation.URL, "application/json", bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("POST %s error: %v", delegation.URL, err)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode/100 != 2 {
		b, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return err
		}

		msg := fmt.Sprintf("Delegation response status %d, body: %s", rsp.StatusCode, string(b))
		return errors.New(msg)
	}

	return nil
}
