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

package websocket

import (
	"sync"
)

//ISession is the interface for session.
type ISession interface {
	OnClosed()
	OnStart(sSessionID string)
	OnReceive(datapacket IDataPacket) bool
	Send(datapacket IDataPacket)
	GetSessionID() string
}

var mSessionList *SessionList

// SessionList is the type for session list.
type SessionList struct {
	sync.RWMutex
	mapOnlineList map[string]ISession
}

//GetSessionList get web client session list
func GetSessionList() *SessionList {
	if mSessionList == nil {
		mSessionList = newSessionList()
	}
	return mSessionList
}

//newSessionList create a new web client session list
func newSessionList() *SessionList {
	return &SessionList{
		mapOnlineList: make(map[string]ISession),
	}
}

//addOnlineSession add one online session
func (sl *SessionList) addOnlineSession(iSession ISession) {
	if iSession.GetSessionID() != "" {
		sl.Lock()
		defer sl.Unlock()
		sl.mapOnlineList[iSession.GetSessionID()] = iSession
	}
}

//removeSession remove one session from web client session list
func (sl *SessionList) removeSession(sSessionID string) (err error) {
	if sSessionID != "" {
		sl.Lock()
		defer sl.Unlock()
		delete(sl.mapOnlineList, sSessionID)
	}
	return err
}

//GetSession get one session from web client session list by sessionid
func (sl *SessionList) GetSession(sSessionID string) ISession {
	sl.RLock()
	defer sl.RUnlock()
	iSession, bOk := sl.mapOnlineList[sSessionID]
	if !bOk {
		return nil
	}
	return iSession
}
