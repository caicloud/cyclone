package v1alpha1

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/caicloud/nirvana/log"
	"github.com/gorilla/websocket"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/biz/accelerator"
	"github.com/caicloud/cyclone/pkg/server/biz/stream"
	"github.com/caicloud/cyclone/pkg/server/biz/utils"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1/sorter"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
	fileutil "github.com/caicloud/cyclone/pkg/util/file"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	websocketutil "github.com/caicloud/cyclone/pkg/util/websocket"
)

// CreateWorkflowRun ...
func CreateWorkflowRun(ctx context.Context, project, workflow, tenant string, wfr *v1alpha1.WorkflowRun) (*v1alpha1.WorkflowRun, error) {
	modifiers := []CreationModifier{GenerateNameModifier, InjectProjectLabelModifier, InjectWorkflowLabelModifier, InjectWorkflowOwnerRefModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, project, workflow, wfr)
		if err != nil {
			return nil, err
		}
	}

	if wfr.Spec.WorkflowRef == nil {
		wfr.Spec.WorkflowRef = workflowReference(tenant, workflow)
	}

	accelerator.NewAccelerator(tenant, project, wfr).Accelerate()
	return handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Create(wfr)
}

// workflowReference returns a workflowRef
func workflowReference(tenant, workflow string) *core_v1.ObjectReference {
	return &core_v1.ObjectReference{
		APIVersion: v1alpha1.APIVersion,
		Kind:       reflect.TypeOf(v1alpha1.Workflow{}).Name(),
		Namespace:  common.TenantNamespace(tenant),
		Name:       workflow,
	}
}

// ListWorkflowRuns ...
func ListWorkflowRuns(ctx context.Context, project, workflow, tenant string, query *types.QueryParams) (*types.ListResponse, error) {
	workflowruns, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: meta.ProjectSelector(project) + "," + meta.WorkflowSelector(workflow),
	})
	if err != nil {
		log.Errorf("Get workflowruns from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, err
	}

	items := workflowruns.Items
	var results []v1alpha1.WorkflowRun
	if query.Filter == "" {
		results = items
	} else {
		results, err = filterWorkflowRuns(items, query.Filter)
		if err != nil {
			return nil, err
		}
	}

	size := int64(len(results))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.WorkflowRun{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	if query.Sort {
		sort.Sort(sorter.NewWorkflowRunSorter(results, query.Ascending))
	}

	return types.NewListResponse(int(size), results[query.Start:end]), nil
}

func filterWorkflowRuns(wfrs []v1alpha1.WorkflowRun, filter string) ([]v1alpha1.WorkflowRun, error) {
	if filter == "" {
		return wfrs, nil
	}

	var results []v1alpha1.WorkflowRun
	// Support multiple filters rules, separated with comma.
	filterParts := strings.Split(filter, ",")
	filters := make(map[string]string)
	for _, part := range filterParts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			return nil, cerr.ErrorQueryParamNotCorrect.Error(filter)
		}

		filters[kv[0]] = strings.ToLower(kv[1])
	}

	var selected bool
	for _, wfr := range wfrs {
		selected = true
		for key, value := range filters {
			switch key {
			case "name":
				if !strings.Contains(wfr.Name, value) {
					selected = false
				}
			case "alias":
				if wfr.Annotations != nil {
					if alias, ok := wfr.Annotations[meta.AnnotationAlias]; ok {
						if strings.Contains(alias, value) {
							continue
						}
					}
				}
				selected = false
			case "status":
				if !strings.EqualFold(string(wfr.Status.Overall.Phase), value) {
					selected = false
				}
			case "trigger":
				if wfr.Annotations != nil {
					if trigger, ok := wfr.Annotations[meta.AnnotationWorkflowRunTrigger]; ok {
						if strings.Contains(trigger, value) {
							continue
						}
					}
				}
				selected = false
			}
		}

		if selected {
			results = append(results, wfr)
		}
	}

	return results, nil
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
		newWfr.Annotations = utils.MergeMap(wfr.Annotations, newWfr.Annotations)
		newWfr.Labels = utils.MergeMap(wfr.Labels, newWfr.Labels)
		if newWfr.Spec.WorkflowRef == nil {
			newWfr.Spec.WorkflowRef = workflowReference(tenant, workflow)
		}

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
	err := deleteCollections(tenant, project, workflow, workflowrun)
	if err != nil {
		return err
	}

	err = handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Delete(workflowrun, nil)
	return cerr.ConvertK8sError(err)
}

// StopWorkflowRun stops a WorkflowRun.
func StopWorkflowRun(ctx context.Context, project, workflow, workflowrun, tenant string) (*v1alpha1.WorkflowRun, error) {
	wfr, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Get(workflowrun, metav1.GetOptions{})
	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	// If wfr already in terminated state, skip it
	if wfr.Status.Overall.Phase == v1alpha1.StatusSucceeded ||
		wfr.Status.Overall.Phase == v1alpha1.StatusFailed ||
		wfr.Status.Overall.Phase == v1alpha1.StatusCancelled {
		return wfr, nil
	}

	data, err := handler.BuildWfrStatusPatch(v1alpha1.StatusCancelled)
	if err != nil {
		log.Errorf("Stop WorkflowRun %s error %s", workflowrun, err)
		return nil, err
	}

	wfr, err = handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Patch(workflowrun, k8s_types.JSONPatchType, data)

	return wfr, cerr.ConvertK8sError(err)
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

// ResumeWorkflowRun updates the workflowrun overall status to Running.
func ResumeWorkflowRun(ctx context.Context, project, workflow, workflowrun, tenant string) (*v1alpha1.WorkflowRun, error) {
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
		log.Errorf("get wfr %s/%s error %s", namespace, workflowrun, err)
		return err
	}

	// get tenant, project, workflow from workflowrun
	tenant := common.NamespaceTenant(namespace)
	var project, workflow string
	if wfr.Labels != nil {
		project = wfr.Labels[meta.LabelProjectName]
		workflow = wfr.Labels[meta.LabelWorkflowName]
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
		log.Infof("log file %s already exists, append logs", logFilePath)
	}

	file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

	logFolder, err := getLogFolder(tenant, project, workflow, workflowrun)
	if err != nil {
		return err
	}
	prefix := fmt.Sprintf("%s_", stage)
	exclusions := []string{fmt.Sprintf("%s_csc-co", stage), fmt.Sprintf("%s_csc-dind", stage)}
	folderReader := stream.NewFolderReader(logFolder, prefix, exclusions, time.Second*10)
	defer folderReader.Close()
	var line []byte
	for {
		select {
		case <-pingTicker.C:
			err := ws.SetWriteDeadline(time.Now().Add(websocketutil.WriteWait))
			if err != nil {
				log.Warningf("set write deadline error:%v", err)
			}
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
					return nil
				}
				return err
			}
		case <-sendTicker.C:
			// With buf.ReadBytes, when err is not nil (often io.EOF), line is not guaranteed to be empty,
			// it holds data before the error occurs.
			line, err = folderReader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				err = ws.WriteMessage(websocket.CloseMessage, []byte("Interval error happens, TERMINATE"))
				if err != nil {
					log.Warningf("write close message error:%v", err)
				}
				break
			}

			if len(line) > 0 {
				err = ws.WriteMessage(websocket.TextMessage, line)
				if err != nil {
					if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
						return nil
					}
					return err
				}
				err = ws.SetWriteDeadline(time.Now().Add(websocketutil.WriteWait))
				if err != nil {
					log.Warningf("set write deadline error:%v", err)
				}
			}
		}
	}
}

// GetContainerLogs handles the request to get container logs, only supports finished stage records.
func GetContainerLogs(ctx context.Context, project, workflow, workflowrun, tenant, stage, container string, download bool) (io.ReadCloser, map[string]string, error) {
	headers := make(map[string]string)
	headers[httputil.HeaderContentType] = "text/plain"
	if download {
		logFileName := fmt.Sprintf("%s-%s-%s-log.txt", workflowrun, stage, container)
		headers["Content-Disposition"] = fmt.Sprintf("attachment; filename=%s", logFileName)
	}

	logFolder, _ := getLogFolder(tenant, project, workflow, workflowrun)
	prefix := fmt.Sprintf("%s_", stage)
	exclusions := []string{fmt.Sprintf("%s_csc-co", stage), fmt.Sprintf("%s_csc-dind", stage)}
	folderReader := stream.NewFolderReader(logFolder, prefix, exclusions, 0)

	return folderReader, headers, nil
}
