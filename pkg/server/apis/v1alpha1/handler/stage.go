package handler

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/caicloud/nirvana/log"
	"github.com/gorilla/websocket"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"bufio"
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
	fileutil "github.com/caicloud/cyclone/pkg/util/file"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	websocketutil "github.com/caicloud/cyclone/pkg/util/websocket"
)

// POST /apis/v1alpha1/stages/
func CreateStage(ctx context.Context) (*v1alpha1.Stage, error) {
	s := &v1alpha1.Stage{}
	err := contextutil.GetJsonPayload(ctx, s)
	if err != nil {
		return nil, err
	}

	return k8sClient.CycloneV1alpha1().Stages(s.Namespace).Create(s)
}

// GET /apis/v1alpha1/Stages/
func ListStages(ctx context.Context, namespace string) (*v1alpha1.StageList, error) {
	return k8sClient.CycloneV1alpha1().Stages(namespace).List(metav1.ListOptions{})
}

// GET /apis/v1alpha1/stages/{stage}
func GetStage(ctx context.Context, name, namespace string) (*v1alpha1.Stage, error) {
	return k8sClient.CycloneV1alpha1().Stages(namespace).Get(name, metav1.GetOptions{})
}

// PUT /apis/v1alpha1/stages/{stage}
func UpdateStage(ctx context.Context, name string) (*v1alpha1.Stage, error) {
	s := &v1alpha1.Stage{}
	err := contextutil.GetJsonPayload(ctx, s)
	if err != nil {
		return nil, err
	}

	if name != s.Name {
		return nil, cerr.ErrorValidationFailed.Error("Name", "Stage name inconsistent between body and path.")
	}

	return k8sClient.CycloneV1alpha1().Stages(s.Namespace).Update(s)
}

// DELETE /apis/v1alpha1/stages/{stage}
func DeleteStage(ctx context.Context, name, namespace string) error {
	return k8sClient.CycloneV1alpha1().Stages(namespace).Delete(name, nil)
}

// GET /workflowruns/{workflowrun-name}/stages/{stage-name}/streamlogs?container-name=c0
// ReceiveContainerLogStream receives real-time log of container within workflowrun stage.
func ReceiveContainerLogStream(ctx context.Context, workflowrun, stage, container, namespace string) error {
	request := contextutil.GetHttpRequest(ctx)
	writer := contextutil.GetHttpResponseWriter(ctx)

	//upgrade HTTP rest API --> socket connection
	ws, err := websocketutil.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Errorf(fmt.Sprintf("Unable to upgrade websocket for err: %v", err))
		return cerr.ErrorUnknownInternal.Error(err)
	}
	defer ws.Close()

	if err := receiveContainerLogStream(workflowrun, stage, container, namespace, ws); err != nil {
		log.Errorf("Fail to receive log stream for workflow(%s):stage(%s):container(%s) : %s",
			workflowrun, stage, container, err.Error())
		return cerr.ErrorUnknownInternal.Error(err)
	}

	return nil
}

// receiveContainerLogStream receives the log stream for
// one stage of the workflowrun, and stores it into log files.
func receiveContainerLogStream(workflowrun, stage, container, namespace string, ws *websocket.Conn) error {
	logFolder, err := getLogFolder(workflowrun, stage, namespace)
	if err != nil {
		log.Errorf("get log folder failed: %v", err)
		return err
	}

	// create log folders.
	fileutil.CreateDirectory(logFolder)

	logFilePath, err := getLogFilePath(workflowrun, stage, container, namespace)
	if err != nil {
		log.Errorf("get log path failed: %v", err)
		return err
	}

	if fileutil.FileExists(logFilePath) {
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
func GetContainerLogStream(ctx context.Context, workflowrun, stage, container, namespace string) error {
	request := contextutil.GetHttpRequest(ctx)
	writer := contextutil.GetHttpResponseWriter(ctx)

	//upgrade HTTP rest API --> socket connection
	ws, err := websocketutil.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to upgrade websocket for err: %s", err.Error()))
		return cerr.ErrorUnknownInternal.Error(err.Error())
	}
	defer ws.Close()

	if err := getContainerLogStream(workflowrun, stage, container, namespace, ws); err != nil {
		log.Errorf("Unable to get logstream for %s/%s/%s/%s for err: %s",
			namespace, workflowrun, stage, container, err)
		return cerr.ErrorUnknownInternal.Error(err.Error())
	}

	return nil
}

// getContainerLogStream watches the log files and sends the content to the log stream.
func getContainerLogStream(workflowrun, stage, container, namespace string, ws *websocket.Conn) error {
	logFilePath, err := getLogFilePath(workflowrun, stage, container, namespace)
	if err != nil {
		log.Errorf("get log path failed: %v", err)
		return err
	}

	if !fileutil.FileExists(logFilePath) {
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

	return nil
}

// GetContainerLogs handles the request to get container logs, only supports finished stage records.
func GetContainerLogs(ctx context.Context, workflowrun, stage, container, namespace string, download bool) ([]byte, map[string]string, error) {
	logs, err := getContainerLogs(workflowrun, stage, container, namespace)
	if err != nil {
		return nil, nil, err
	}

	headers := make(map[string]string)
	headers[httputil.HEADER_ContentType] = "text/plain"
	if download {
		logFileName := fmt.Sprintf("%s-%s-%s-log.txt", workflowrun, stage, container)
		headers["Content-Disposition"] = fmt.Sprintf("attachment; filename=%s", logFileName)
	}

	return []byte(logs), headers, nil
}

// getContainerLogs gets the stage container logs.
func getContainerLogs(workflowrun, stage, container, namespace string) (string, error) {
	logFilePath, err := getLogFilePath(workflowrun, stage, container, namespace)
	if err != nil {
		return "", err
	}

	// Check the existence of the log file for this stage. If does not exist, return error when stage is success,
	// otherwise directly return the got logs as stage is failed or aborted.
	if !fileutil.FileExists(logFilePath) {
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
