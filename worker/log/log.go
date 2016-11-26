/*
Copyright 2016 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/executil"
	"github.com/caicloud/cyclone/pkg/filebuffer"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/pathutil"
	"golang.org/x/net/websocket"
)

const (
	// LogSpecialMark represents special mark.
	LogSpecialMark string = "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
	// LogReplaceMark represents replace mark.
	LogReplaceMark string = "--->"
	// FileBufferSize is the default size of file buffers.
	FileBufferSize = 64 * 1024 * 1024
)

var (
	// Output is used to collect log information.
	Output filebuffer.FileBuffer
)

// StepEvent is information about step event name in creating versions
type StepEvent string

const (
	CloneRepository StepEvent = "clone repository"
	CreateTag       StepEvent = "create tag"
	BuildImage      StepEvent = "Build image"
	PushImage       StepEvent = "Push image"
	Integration     StepEvent = "Integration"
	PreBuild        StepEvent = "Pre Build"
	PostBuild       StepEvent = "Post Build"
	Deploy          StepEvent = "Deploy application"
	ApplyResource   StepEvent = "Apply Resource"
	ParseYaml       StepEvent = "Parse Yaml"
)

// StepState is information about step event's state
type StepState string

const (
	Start  StepState = "start"
	Stop   StepState = "stop"
	Finish StepState = "finish"
)

// InsertStepLog inserts the step information into the log file.
func InsertStepLog(event *api.Event, stepevent StepEvent, state StepState, err error) {
	var stepLog string
	if nil == err {
		stepLog = fmt.Sprintf("step: %s state: %s", stepevent, state)
	} else {
		stepLog = fmt.Sprintf("step: %s state: %s Error: %v", stepevent, state, err)
	}

	fmt.Fprintf(Output, "%s\n", stepLog)
}

// RemoveLogFile is used to delete the log file.
func RemoveLogFile(filepath string) error {
	// File does or  does not exist.
	_, err := os.Stat(filepath)
	if err != nil && os.IsNotExist(err) {
		return err
	}

	filename := path.Base(filepath)
	dir := path.Dir(filepath)

	args := []string{"-f", filename}
	_, err = executil.RunInDir(dir, "rm", args...)

	if err != nil {
		return fmt.Errorf("Error when removing log file  %s", filename)
	}
	log.InfoWithFields("Successfully removed log file", log.Fields{"filename": filename})
	return err
}

// CreateFileBuffer func that create a file buffer for storage log
func CreateFileBuffer(eventID api.EventID) error {
	var path string
	path = fmt.Sprintf("/logs/%s", eventID)
	err := pathutil.EnsureParentDir(path, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := osutil.OpenFile(path, os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	Output = filebuffer.NewFileBuffer(FileBufferSize, file)
	return nil
}

var (
	ws                  *websocket.Conn
	lockFileWatchSwitch sync.RWMutex
	watchLogFileSwitch  map[string]bool
)

// DialLogServer dial and connect to the log server
// e.g
// origin "http://120.26.103.63/"
// url "ws://120.26.103.63:8000/ws"
func DialLogServer(url string) error {
	addr := strings.Split(url, "/")[2]
	origin := "http://" + addr + "/"
	log.Infof("Dail to log server: url(%s), origin(%s)", url, origin)

	var err error
	ws, err = websocket.Dial(url, "", origin)
	return err
}

// Disconnect dicconnect websocket from log server
func Disconnect() {
	ws.Close()
}

// HeatBeatPacket is the type for heart_beat packet.
type HeatBeatPacket struct {
	Action string `json:"action"`
	Id     string `json:"id"`
}

// SendHeartBeat send heat beat packet to log server per 30 seconds.
func SendHeartBeat() {
	id := 0

	for {
		if nil == ws {
			return
		}

		structData := &HeatBeatPacket{
			Action: "heart_beat",
			Id:     fmt.Sprintf("%d", id),
		}
		jsonData, _ := json.Marshal(structData)

		if _, err := ws.Write(jsonData); err != nil {
			log.Errorf("Send heart beat to server err: %v", err)
		}

		id++
		time.Sleep(time.Second * 30)
	}
}

// SetWatchLogFileSwitch set swicth of watch log file
func SetWatchLogFileSwitch(filePath string, watchSwitch bool) {
	lockFileWatchSwitch.Lock()
	defer lockFileWatchSwitch.Unlock()
	if nil == watchLogFileSwitch {
		watchLogFileSwitch = make(map[string]bool)
	}
	watchLogFileSwitch[filePath] = watchSwitch
}

// GetWatchLogFileSwitch set swicth of watch log file
func GetWatchLogFileSwitch(filePath string) bool {
	lockFileWatchSwitch.RLock()
	defer lockFileWatchSwitch.RUnlock()
	bEnable, bFound := watchLogFileSwitch[filePath]
	if bFound {
		return bEnable
	}
	return false

}

// WatchLogFile watch the log and prouce one line to kafka topic per 200ms
func WatchLogFile(filePath string, topic string, ch chan interface{}) {
	var logFile *os.File
	var err error

	SetWatchLogFileSwitch(filePath, true)
	for {
		if false == GetWatchLogFileSwitch(filePath) {
			close(ch)
			return
		}

		logFile, err = os.Open(filePath)
		if nil != err {
			log.Debugf("open log file haven't create: %v", err)
			// wait for log file create
			time.Sleep(time.Second)
			continue
		}
		break
	}
	defer func() {
		logFile.Close()
	}()

	buf := bufio.NewReader(logFile)
	// read log file 30 lines for one time or read to the end
	for {
		if false == GetWatchLogFileSwitch(filePath) {
			var lines = ""
			for {
				line, errRead := buf.ReadString('\n')
				if errRead != nil {
					if errRead == io.EOF {
						break
					}
					log.Errorf("watch log file errs: %v", errRead)
					close(ch)
					return
				}
				line = strings.TrimSpace(line)
				lines += (line + "\r\n")
			}
			if len(lines) != 0 {
				pushLog(topic, lines)
			}
			close(ch)
			return
		}

		var lines string
		for i := 0; i < 30; i++ {
			line, errRead := buf.ReadString('\n')
			if errRead != nil {
				if errRead == io.EOF {
					break
				}
				log.Errorf("watch log file err: %v", errRead)
				close(ch)
				return
			}
			line = strings.TrimSpace(line)
			lines += (line + "\r\n")
		}
		if len(lines) != 0 {
			pushLog(topic, lines)
		}
		time.Sleep(time.Millisecond * 100)
	}
}

// PushLogPacket is the type for push_log packet.
type PushLogPacket struct {
	Action string `json:"action"`
	Topic  string `json:"topic"`
	Log    string `json:"log"`
}

// pushLog push log to log server
func pushLog(topic string, slog string) {
	if nil == ws {
		return
	}

	structData := &PushLogPacket{
		Action: "worker_push_log",
		Topic:  topic,
		Log:    slog,
	}
	jsonData, _ := json.Marshal(structData)

	if _, err := ws.Write(jsonData); err != nil {
		log.Errorf("Push log to server err: %v", err)
	}
}
