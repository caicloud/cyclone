package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/caicloud/nirvana/log"
	"github.com/gorilla/websocket"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
	fileutil "github.com/caicloud/cyclone/pkg/util/file"
	websocketutil "github.com/caicloud/cyclone/pkg/util/websocket"
)

// POST /apis/v1alpha1/stages/
// X-Tenant: any
func CreateStage(ctx context.Context) (*v1alpha1.Stage, error) {
	s := &v1alpha1.Stage{}
	err := contextutil.GetJsonPayload(ctx, s)
	if err != nil {
		return nil, err
	}

	return k8sClient.CycloneV1alpha1().Stages(s.Namespace).Create(s)
}

// POST /apis/v1alpha1/stages/{stage-name}
// X-Tenant: any
func GetStage(ctx context.Context, name, namespace string) (*v1alpha1.Stage, error) {
	return k8sClient.CycloneV1alpha1().Stages(namespace).Get(name, metav1.GetOptions{})
}

// GET /workflowruns/{workflowrun-name}/stages/{stage-name}/streamlogs?container-name=c0
// ReceiveContainerLogStream receives real-time log of container within workflowrun stage.
func ReceiveContainerLogStream(ctx context.Context, workflowrun, stage, container string) error {
	request := contextutil.GetHttpRequest(ctx)
	writer := contextutil.GetHttpResponseWriter(ctx)

	//upgrade HTTP rest API --> socket connection
	ws, err := websocketutil.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Errorf(fmt.Sprintf("Unable to upgrade websocket for err: %v", err))
		return cerr.ErrorUnknownInternal.Error(err)
	}
	defer ws.Close()

	if err := receiveContainerLogStream(workflowrun, stage, container, ws); err != nil {
		log.Errorf("Fail to receive log stream for workflow(%s):stage(%s):container(%s) : %s",
			workflowrun, stage, container, err.Error())
		return cerr.ErrorUnknownInternal.Error(err)
	}

	return nil
}

// receiveContainerLogStream receives the log stream for
// one stage of the workflowrun, and stores it into log files.
func receiveContainerLogStream(workflowrun, stage, container string, ws *websocket.Conn) error {
	logFolder, err := getLogFolder(workflowrun, stage)
	if err != nil {
		log.Errorf("get log folder failed: %v", err)
		return err
	}

	// create log folders.
	fileutil.CreateDirectory(logFolder)

	logFilePath, err := getLogFilePath(workflowrun, stage, container)
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
