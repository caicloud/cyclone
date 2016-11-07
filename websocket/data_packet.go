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

//IDataPacket is the interface for data packet.
type IDataPacket interface {
	GetLength() int
	GetData() []byte
	GetSendTo() string
	GetReceiveFrom() string
}

// DataPacket is the type for data packet.
type DataPacket struct {
	//data
	byrFrame []byte
	//length of data
	nFrameLen int
	//send to which session
	sSendTo string
	//receiver from which session
	sReceiveFrom string
}

// GetLength gets the length of data.
func (p *DataPacket) GetLength() int {
	if nil == p.byrFrame {
		return 0
	} else if len(p.byrFrame) > p.nFrameLen {
		return p.nFrameLen
	} else {
		return len(p.byrFrame)
	}
}

// GetData gets date from data packet.
func (p *DataPacket) GetData() []byte {
	return p.byrFrame
}

// SetLength sets length of data.
func (p *DataPacket) SetLength(nLen int) {
	p.nFrameLen = nLen
}

// SetData sets data in data packet.
func (p *DataPacket) SetData(byrFrame []byte) {
	p.byrFrame = byrFrame
}

// GetSendTo gets the session id of the date need send to.
func (p *DataPacket) GetSendTo() string {
	return p.sSendTo
}

// SetSendTo sets the session id of the date need send to.
func (p *DataPacket) SetSendTo(sSendTo string) {
	p.sSendTo = sSendTo
}

// GetReceiveFrom gets the session id of the date receive from.
func (p *DataPacket) GetReceiveFrom() string {
	return p.sReceiveFrom
}

// SetReceiveFrom sets the session id of the date receive from.
func (p *DataPacket) SetReceiveFrom(sReceiveFrom string) {
	p.sReceiveFrom = sReceiveFrom
}
