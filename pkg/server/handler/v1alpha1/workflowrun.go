package v1alpha1

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/caicloud/nirvana/log"
	"github.com/gorilla/websocket"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
	fileutil "github.com/caicloud/cyclone/pkg/util/file"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	websocketutil "github.com/caicloud/cyclone/pkg/util/websocket"
)

// CreateWorkflowRun ...
func CreateWorkflowRun(ctx context.Context, project, workflow, tenant string, wfr *v1alpha1.WorkflowRun) (*v1alpha1.WorkflowRun, error) {
	modifiers := []CreationModifier{GenerateNameModifier, InjectProjectLabelModifier, WorkflowRunModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, project, workflow, wfr)
		if err != nil {
			return nil, err
		}
	}

	injectWfRef(tenant, workflow, wfr)
	return handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Create(wfr)
}

// injectWfRef injects workflowRef if it is nil
func injectWfRef(tenant, workflow string, wfr *v1alpha1.WorkflowRun) {
	if wfr.Spec.WorkflowRef == nil {
		wfr.Spec.WorkflowRef = &core_v1.ObjectReference{
			Namespace: common.TenantNamespace(tenant),
			Name:      workflow,
			Kind:      "workflow.cyclone.io",
		}
	}
}

// ListWorkflowRuns ...
func ListWorkflowRuns(ctx context.Context, project, workflow, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	workflowruns, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: common.ProjectSelector(project) + "," + common.WorkflowSelector(workflow),
	})
	if err != nil {
		log.Errorf("Get workflowruns from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, err
	}

	items := workflowruns.Items
	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.WorkflowRun{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[pagination.Start:end]), nil
}

// GetWorkflowRun ...
func GetWorkflowRun(ctx context.Context, project, workflow, workflowrun, tenant string) (*v1alpha1.WorkflowRun, error) {
	wfr, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Get(workflowrun, metav1.GetOptions{})

	return wfr, cerr.ConvertK8sError(err)
}

// UpdateWorkflowRun ...
func UpdateWorkflowRun(ctx context.Context, project, workflow, workflowrun, tenant string, wfr *v1alpha1.WorkflowRun) (*v1alpha1.WorkflowRun, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Get(workflowrun, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newWfr := origin.DeepCopy()
		newWfr.Spec = wfr.Spec
		newWfr.Annotations = MergeMap(wfr.Annotations, newWfr.Annotations)
		newWfr.Labels = MergeMap(wfr.Labels, newWfr.Labels)
		injectWfRef(tenant, workflow, wfr)
		_, err = handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Update(newWfr)
		return err
	})

	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return wfr, nil
}

// DeleteWorkflowRun ...
func DeleteWorkflowRun(ctx context.Context, project, workflow, workflowrun, tenant string) error {
	err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Delete(workflowrun, nil)

	return cerr.ConvertK8sError(err)
}

// PauseWorkflowRun updates the workflowrun overall status to Waiting.
func PauseWorkflowRun(ctx context.Context, project, workflow, workflowrun, tenant string) (*v1alpha1.WorkflowRun, error) {
	data, err := handler.BuildWfrStatusPatch(v1alpha1.StatusWaiting)
	if err != nil {
		log.Errorf("pause workflowrun %s error %s", workflowrun, err)
		return nil, err
	}

	wfr, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Patch(workflowrun, k8s_types.JSONPatchType, data)

	return wfr, cerr.ConvertK8sError(err)
}

// ContinueWorkflowRun updates the workflowrun overall status to Running.
func ContinueWorkflowRun(ctx context.Context, project, workflow, workflowrun, tenant string) (*v1alpha1.WorkflowRun, error) {
	data, err := handler.BuildWfrStatusPatch(v1alpha1.StatusRunning)
	if err != nil {
		log.Errorf("continue workflowrun %s error %s", workflowrun, err)
		return nil, err
	}

	wfr, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Patch(workflowrun, k8s_types.JSONPatchType, data)

	return wfr, cerr.ConvertK8sError(err)
}

// ReceiveContainerLogStream receives real-time log of container within workflowrun stage.
func ReceiveContainerLogStream(ctx context.Context, workflowrun, namespace, stage, container string) error {
	// get workflowrun
	wfr, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Get(workflowrun, metav1.GetOptions{})
	if err != nil {
		log.Errorf("get wfr %s error %s", wfr, err)
		return err
	}

	// get tenant, project, workflow from workflowrun
	tenant := common.NamespaceTenant(namespace)
	var project, workflow string
	if wfr.Labels != nil {
		project = wfr.Labels[common.LabelProjectName]
		workflow = wfr.Labels[common.LabelWorkflowName]
	}
	if project == "" || workflow == "" {
		return fmt.Errorf("failed to get project or workflow from workflowrun labels")
	}

	request := contextutil.GetHTTPRequest(ctx)
	writer := contextutil.GetHTTPResponseWriter(ctx)

	//upgrade HTTP rest API --> socket connection
	ws, err := websocketutil.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Errorf(fmt.Sprintf("Unable to upgrade websocket for err: %v", err))
		return cerr.ErrorUnknownInternal.Error(err)
	}
	defer ws.Close()

	if err := receiveContainerLogStream(tenant, project, workflow, workflowrun, stage, container, ws); err != nil {
		log.Errorf("Fail to receive log stream for workflow(%s):stage(%s):container(%s) : %s",
			workflowrun, stage, container, err.Error())
		return cerr.ErrorUnknownInternal.Error(err)
	}

	return nil
}

// receiveContainerLogStream receives the log stream for
// one stage of the workflowrun, and stores it into log files.
func receiveContainerLogStream(tenant, project, workflow, workflowrun, stage, container string, ws *websocket.Conn) error {
	logFolder, err := getLogFolder(tenant, project, workflow, workflowrun)
	if err != nil {
		log.Errorf("get log folder failed: %v", err)
		return err
	}

	// create log folders.
	fileutil.CreateDirectory(logFolder)

	logFilePath, err := getLogFilePath(tenant, project, workflow, workflowrun, stage, container)
	if err != nil {
		log.Errorf("get log path failed: %v", err)
		return err
	}

	if fileutil.Exists(logFilePath) {
		return fmt.Errorf("log file %s already exists", logFilePath)
	}

	file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("fail to open the log file %s as %v", logFilePath, err)
		return err
	}
	defer file.Close()

	var message []byte
	for {
		_, message, err = ws.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				return nil
			}

			log.Infoln(err)
			return err
		}
		_, err = file.Write(message)
		if err != nil {
			return err
		}
	}
}

// GetContainerLogStream gets real-time log of container within stage.
func GetContainerLogStream(ctx context.Context, project, workflow, workflowrun, tenant, stage, container string) error {
	request := contextutil.GetHTTPRequest(ctx)
	writer := contextutil.GetHTTPResponseWriter(ctx)

	//upgrade HTTP rest API --> socket connection
	ws, err := websocketutil.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to upgrade websocket for err: %s", err.Error()))
		return cerr.ErrorUnknownInternal.Error(err.Error())
	}
	defer ws.Close()

	if err := getContainerLogStream(tenant, project, workflow, workflowrun, stage, container, ws); err != nil {
		log.Errorf("Unable to get logstream for %s/%s/%s for err: %s",
			workflowrun, stage, container, err)
		return cerr.ErrorUnknownInternal.Error(err.Error())
	}

	return nil
}

// getContainerLogStream watches the log files and sends the content to the log stream.
func getContainerLogStream(tenant, project, workflow, workflowrun, stage, container string, ws *websocket.Conn) error {
	logFilePath, err := getLogFilePath(tenant, project, workflow, workflowrun, stage, container)
	if err != nil {
		log.Errorf("get log path failed: %v", err)
		return err
	}

	if !fileutil.Exists(logFilePath) {
		return fmt.Errorf("log file %s does not exist", logFilePath)
	}

	pingTicker := time.NewTicker(websocketutil.PingPeriod)
	sendTicker := time.NewTicker(10 * time.Millisecond)
	file, err := os.Open(logFilePath)
	if err != nil {
		log.Errorf("fail to open the log file %s as %s", logFilePath, err.Error())
		return err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	var line []byte
	for {
		select {
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(websocketutil.WriteWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
					return nil
				}
				return err
			}
		case <-sendTicker.C:
			line, err = buf.ReadBytes('\n')
			if err == io.EOF {
				continue
			}

			if err != nil {
				ws.WriteMessage(websocket.CloseMessage, []byte("Interval error happens, TERMINATE"))
				break
			}

			err = ws.WriteMessage(websocket.TextMessage, line)
			if err != nil {
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
					return nil
				}
				return err
			}
			ws.SetWriteDeadline(time.Now().Add(websocketutil.WriteWait))
		}
	}

}

// GetContainerLogs handles the request to get container logs, only supports finished stage records.
func GetContainerLogs(ctx context.Context, project, workflow, workflowrun, tenant, stage, container string, download bool) ([]byte, map[string]string, error) {
	logs, err := getContainerLogs(tenant, project, workflow, workflowrun, stage, container)
	if err != nil {
		return nil, nil, err
	}

	headers := make(map[string]string)
	headers[httputil.HeaderContentType] = "text/plain"
	if download {
		logFileName := fmt.Sprintf("%s-%s-%s-log.txt", workflowrun, stage, container)
		headers["Content-Disposition"] = fmt.Sprintf("attachment; filename=%s", logFileName)
	}

	return []byte(logs), headers, nil
}

// getContainerLogs gets the stage container logs.
func getContainerLogs(tenant, project, workflow, workflowrun, stage, container string) (string, error) {
	logFilePath, err := getLogFilePath(tenant, project, workflow, workflowrun, stage, container)
	if err != nil {
		return "", err
	}

	// Check the existence of the log file for this stage. If does not exist, return error when stage is success,
	// otherwise directly return the got logs as stage is failed or aborted.
	if !fileutil.Exists(logFilePath) {
		log.Errorf("log file %s does not exist", logFilePath)
		return "", fmt.Errorf("log file for stage %s does not exist", stage)
	}

	// TODO (robin) Read the whole file, need to consider the memory consumption when the log file is too huge.
	log, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		return "", err
	}

	return string(log), nil
}
