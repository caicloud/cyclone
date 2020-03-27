package v1alpha1

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
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
	wfcommon "github.com/caicloud/cyclone/pkg/workflow/common"
)

const (
	// stageArtifactFormFileKey is the form file key to receive stage artifact
	stageArtifactFormFileKey = "file"
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

	size := uint64(len(results))
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
			case queryFilterName:
				if !strings.Contains(wfr.Name, value) {
					selected = false
				}
			case queryFilterAlias:
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

	defer func() {
		if err := ws.Close(); err != nil {
			log.Errorf("Fail to close websocket as: %v", err)
		}
	}()

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

	defer func() {
		if err := file.Close(); err != nil {
			log.Errorf("Fail to close file as: %v", err)
		}
	}()

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

// GetContainerLogStream gets real-time logs of a certain stage.
func GetContainerLogStream(ctx context.Context, project, workflow, workflowrun, tenant, stage, qt string) error {
	// If there is no tenant info in Header, we will get it from query, since front end is hard to
	// pass information in Header while using Websockets
	// More info: https://stackoverflow.com/questions/4361173/http-headers-in-websockets-client-api
	if tenant == "" {
		tenant = qt
	}

	request := contextutil.GetHTTPRequest(ctx)
	writer := contextutil.GetHTTPResponseWriter(ctx)

	//upgrade HTTP rest API --> socket connection
	ws, err := websocketutil.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to upgrade websocket for err: %s", err.Error()))
		return cerr.ErrorUnknownInternal.Error(err.Error())
	}

	defer func() {
		if err := ws.Close(); err != nil {
			log.Errorf("Fail to close websocket as: %v", err)
		}
	}()

	if err := getContainerLogStream(tenant, project, workflow, workflowrun, stage, ws); err != nil {
		log.Errorf("Unable to get logstream for %s/%s for err: %s", workflowrun, stage, err)
		return cerr.ErrorUnknownInternal.Error(err.Error())
	}

	return nil
}

// getContainerLogStream watches the log files and sends the content to the log stream.
func getContainerLogStream(tenant, project, workflow, workflowrun, stage string, ws *websocket.Conn) error {
	logFolder, err := getLogFolder(tenant, project, workflow, workflowrun)
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("%s_", stage)
	exclusions := []string{fmt.Sprintf("%s_%s", stage, wfcommon.CoordinatorSidecarName), fmt.Sprintf("%s_%s", stage, wfcommon.DockerInDockerSidecarName)}
	ctx, cancel := context.WithCancel(context.Background())
	folderReader := stream.NewFolderReader(logFolder, prefix, exclusions, time.Second*10, cancel)

	defer func() {
		if err := folderReader.Close(); err != nil {
			log.Errorf("Fail to close folder reader as: %v", err)
		}
	}()

	go watchStageTermination(common.TenantNamespace(tenant), workflowrun, stage, cancel)
	err = websocketutil.Write(ws, folderReader, ctx.Done())
	if err != nil {
		log.Error("websocket writer error:", err)
	}
	log.Infof("End of log stream for wfr '%s', stage '%s'", workflowrun, stage)

	return err
}

// watchStageTermination ensures to close the log WebSocket finally when FolderReader can't be terminated (no EOF file in the folder)
func watchStageTermination(namespace, wfrName, stgName string, onTerminatedCallback context.CancelFunc) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		wfr, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Get(wfrName, metav1.GetOptions{})
		if err != nil {
			log.Warningf("Get workflowRun %s error: %v", wfrName, err)
			onTerminatedCallback()
			return
		}

		if wfr.Status.Overall.Phase == v1alpha1.StatusSucceeded ||
			wfr.Status.Overall.Phase == v1alpha1.StatusFailed ||
			wfr.Status.Overall.Phase == v1alpha1.StatusCancelled {
			onTerminatedCallback()
			return
		}

		for stage, status := range wfr.Status.Stages {
			if stage == stgName && (status.Status.Phase == v1alpha1.StatusSucceeded ||
				status.Status.Phase == v1alpha1.StatusFailed ||
				status.Status.Phase == v1alpha1.StatusCancelled) {
				onTerminatedCallback()
				return
			}
		}
	}
}

// GetContainerLogs handles the request to get container logs, only supports finished stage records.
func GetContainerLogs(ctx context.Context, project, workflow, workflowrun, tenant, stage, container string, download bool) (io.ReadCloser, map[string]string, error) {
	headers := make(map[string]string)
	headers[httputil.HeaderContentType] = "text/plain"
	if download {
		logFileName := fmt.Sprintf("%s-%s-log.txt", workflowrun, stage)
		headers["Content-Disposition"] = fmt.Sprintf("attachment; filename=%s", logFileName)
	}

	logFolder, _ := getLogFolder(tenant, project, workflow, workflowrun)
	prefix := fmt.Sprintf("%s_", stage)
	exclusions := []string{fmt.Sprintf("%s_%s", stage, wfcommon.CoordinatorSidecarName), fmt.Sprintf("%s_%s", stage, wfcommon.DockerInDockerSidecarName)}
	folderReader := stream.NewFolderReader(logFolder, prefix, exclusions, 0, nil)

	return folderReader, headers, nil
}

// ReceiveArtifacts receives artifacts produced by workflowrun stage.
func ReceiveArtifacts(ctx context.Context, workflowrun, namespace, stage string) error {
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

	file, fileHeader, err := request.FormFile(stageArtifactFormFileKey)
	if err != nil {
		log.Infof("Form file by key %s error: %v", stageArtifactFormFileKey, err)
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Errorf("Fail to close file as: %v", err)
		}
	}()

	folder, err := getArtifactFolder(tenant, project, workflow, workflowrun, stage)
	if err != nil {
		log.Errorf("Get workflowrun %s artifact folder failed: %v", workflowrun, err)
		return err
	}

	// create artifact folders.
	fileutil.CreateDirectory(folder)
	// copy file
	f, err := os.OpenFile(fmt.Sprintf("%s/%s", folder, fileHeader.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("Fail to close file as: %v", err)
		}
	}()

	_, err = io.Copy(f, file)
	if err != nil {
		log.Infof("copy artifact %s error: %v", fileHeader.Filename, err)
	}

	return err
}

// ListArtifacts handles the request to list artifacts produced in a workflowRun.
func ListArtifacts(ctx context.Context, project, workflow, workflowrun, tenant string) (*types.ListResponse, error) {
	var artifacts []api.StageArtifact

	wf, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Get(workflow, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	for _, stage := range wf.Spec.Stages {
		artifactFolder, _ := getArtifactFolder(tenant, project, workflow, workflowrun, stage.Name)
		artifactFolderInfo, err := os.Stat(artifactFolder)
		if os.IsNotExist(err) {
			continue
		}
		if !artifactFolderInfo.IsDir() {
			continue
		}

		files, err := ioutil.ReadDir(artifactFolder)
		if err != nil {
			log.Infof("Readdir %s error:%v", artifactFolder, err)
			return nil, err
		}

		for _, file := range files {
			// Only support display files, since we can not collect folders, all files will be compressed to a tar file.
			if file.IsDir() {
				continue
			}
			artifact := api.StageArtifact{
				Stage:             stage.Name,
				File:              file.Name(),
				CreationTimestamp: file.ModTime(),
			}

			log.Info(artifact)
			artifacts = append(artifacts, artifact)
		}
	}

	return types.NewListResponse(len(artifacts), artifacts), nil
}

// DownloadArtifact handles the request to download a artifact of a stage produced by a workflowRun.
func DownloadArtifact(ctx context.Context, project, workflow, workflowrun, artifact, tenant, stage string) (io.ReadCloser, map[string]string, error) {
	headers := make(map[string]string)
	headers["Content-Disposition"] = fmt.Sprintf("attachment; filename=%s", artifact)

	artifactFolder, err := getArtifactFolder(tenant, project, workflow, workflowrun, stage)
	if err != nil {
		return nil, nil, err
	}

	artifactFilePath := artifactFolder + "/" + artifact

	artifactFile, err := os.Open(artifactFilePath)
	if err != nil {
		return nil, nil, err
	}

	go func() {
		for range ctx.Done() {
			if err := artifactFile.Close(); err != nil {
				log.Errorf("Fail to close file as: %v", err)
			}
		}
	}()

	return artifactFile, headers, nil
}

// DeleteArtifact handles the request to delete a artifact of a stage produced by a workflowRun.
func DeleteArtifact(ctx context.Context, project, workflow, workflowrun, artifact, tenant, stage string) error {
	artifactFolder, err := getArtifactFolder(tenant, project, workflow, workflowrun, stage)
	if err != nil {
		return err
	}

	artifactFilePath := artifactFolder + "/" + artifact
	return os.RemoveAll(artifactFilePath)
}
