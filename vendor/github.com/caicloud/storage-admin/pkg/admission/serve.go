package admission

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	admsv1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	var (
		ar   admsv1.AdmissionReview
		resp admsv1.AdmissionResponse
	)

	defer r.Body.Close()
	b, e := ioutil.ReadAll(r.Body)
	if e != nil {
		glog.Errorf("serve read request body failed, %v", e)
		return
	}
	glog.Infof("serve start for %v", r.RemoteAddr)
	glog.V(2).Info(string(b))

	if e = json.Unmarshal(b, &ar); e != nil {
		glog.Errorf("serve parse request body failed, %v", e)
		return
	}
	logPath := fmt.Sprintf("serve AdmissionReview %s", admissionReviewToString(&ar))
	glog.Infof("%s start", logPath)

	f, e := s.ruleMap.GetChecker(&ar)
	if e != nil {
		glog.Warningf("%s GetChecker failed, %v", logPath, e)
		ar.Response.Allowed = true
		packAndReturn(&ar, w)
		return
	}
	if e = f(s.kubeClient(), &ar); e != nil {
		glog.Warningf("%s check failed, %v", logPath, e)
		resp.Allowed = false
		resp.Result = &metav1.Status{
			Message: e.Error(),
		}
	} else {
		resp.Allowed = true
	}

	ar.Response = &resp

	if e = packAndReturn(&ar, w); e != nil {
		glog.Warningf("%s packAndReturn failed, %v", logPath, e)
		return
	}
	glog.Infof("%s done", logPath)
}

func packAndReturn(ar *admsv1.AdmissionReview, w io.Writer) error {
	data, e := json.Marshal(ar)
	if e != nil {
		return e
	}
	_, e = w.Write(data)
	return e
}

func admissionReviewCheckerFake(kc kubernetes.Interface, ar *admsv1.AdmissionReview) error {
	logPath := fmt.Sprintf("admissionReviewCheckerFake %v", admissionReviewToString(ar))
	// test print
	b, _ := json.MarshalIndent(ar, "", "    ")
	glog.Infof("%s %s", logPath, string(b))
	return nil
}
