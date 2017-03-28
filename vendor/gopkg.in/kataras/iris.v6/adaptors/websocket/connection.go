package websocket

import (
	"bytes"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/kataras/iris.v6"
)

type (
	connectionValue struct {
		key   []byte
		value interface{}
	}
	// ConnectionValues is the temporary connection's memory store
	ConnectionValues []connectionValue
)

// Set sets a value based on the key
func (r *ConnectionValues) Set(key string, value interface{}) {
	args := *r
	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if string(kv.key) == key {
			kv.value = value
			return
		}
	}

	c := cap(args)
	if c > n {
		args = args[:n+1]
		kv := &args[n]
		kv.key = append(kv.key[:0], key...)
		kv.value = value
		*r = args
		return
	}

	kv := connectionValue{}
	kv.key = append(kv.key[:0], key...)
	kv.value = value
	*r = append(args, kv)
}

// Get returns a value based on its key
func (r *ConnectionValues) Get(key string) interface{} {
	args := *r
	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if string(kv.key) == key {
			return kv.value
		}
	}
	return nil
}

// Reset clears the values
func (r *ConnectionValues) Reset() {
	*r = (*r)[:0]
}

// UnderlineConnection is used for compatible with fasthttp and net/http underline websocket libraries
// we only need ~8 funcs from websocket.Conn so:
type UnderlineConnection interface {
	// SetWriteDeadline sets the write deadline on the underlying network
	// connection. After a write has timed out, the websocket state is corrupt and
	// all future writes will return an error. A zero value for t means writes will
	// not time out.
	SetWriteDeadline(t time.Time) error
	// SetReadDeadline sets the read deadline on the underlying network connection.
	// After a read has timed out, the websocket connection state is corrupt and
	// all future reads will return an error. A zero value for t means reads will
	// not time out.
	SetReadDeadline(t time.Time) error
	// SetReadLimit sets the maximum size for a message read from the peer. If a
	// message exceeds the limit, the connection sends a close frame to the peer
	// and returns ErrReadLimit to the application.
	SetReadLimit(limit int64)
	// SetPongHandler sets the handler for pong messages received from the peer.
	// The appData argument to h is the PONG frame application data. The default
	// pong handler does nothing.
	SetPongHandler(h func(appData string) error)
	// SetPingHandler sets the handler for ping messages received from the peer.
	// The appData argument to h is the PING frame application data. The default
	// ping handler sends a pong to the peer.
	SetPingHandler(h func(appData string) error)
	// WriteControl writes a control message with the given deadline. The allowed
	// message types are CloseMessage, PingMessage and PongMessage.
	WriteControl(messageType int, data []byte, deadline time.Time) error
	// WriteMessage is a helper method for getting a writer using NextWriter,
	// writing the message and closing the writer.
	WriteMessage(messageType int, data []byte) error
	// ReadMessage is a helper method for getting a reader using NextReader and
	// reading from that reader to a buffer.
	ReadMessage() (messageType int, p []byte, err error)
	// NextWriter returns a writer for the next message to send. The writer's Close
	// method flushes the complete message to the network.
	//
	// There can be at most one open writer on a connection. NextWriter closes the
	// previous writer if the application has not already done so.
	NextWriter(messageType int) (io.WriteCloser, error)
	// Close closes the underlying network connection without sending or waiting for a close frame.
	Close() error
}

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------
// -------------------------------Connection implementation-----------------------------
// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

type (
	// DisconnectFunc is the callback which fires when a client/connection closed
	DisconnectFunc func()
	// LeaveRoomFunc is the callback which fires when a client/connection leaves from any room.
	// This is called automatically when client/connection disconnected
	// (because websocket server automatically leaves from all joined rooms)
	LeaveRoomFunc func(roomName string)
	// ErrorFunc is the callback which fires when an error happens
	ErrorFunc (func(string))
	// NativeMessageFunc is the callback for native websocket messages, receives one []byte parameter which is the raw client's message
	NativeMessageFunc func([]byte)
	// MessageFunc is the second argument to the Emitter's Emit functions.
	// A callback which should receives one parameter of type string, int, bool or any valid JSON/Go struct
	MessageFunc interface{}
	// Connection is the front-end API that you will use to communicate with the client side
	Connection interface {
		// Emitter implements EmitMessage & Emit
		Emitter
		// ID returns the connection's identifier
		ID() string

		// Context returns the (upgraded) *iris.Context of this connection
		// avoid using it, you normally don't need it,
		// websocket has everything you need to authenticate the user BUT if it's necessary
		// then  you use it to receive user information, for example: from headers
		Context() *iris.Context

		// OnDisconnect registers a callback which fires when this connection is closed by an error or manual
		OnDisconnect(DisconnectFunc)
		// OnError registers a callback which fires when this connection occurs an error
		OnError(ErrorFunc)
		// EmitError can be used to send a custom error message to the connection
		//
		// It does nothing more than firing the OnError listeners. It doesn't sends anything to the client.
		EmitError(errorMessage string)
		// To defines where server should send a message
		// returns an emitter to send messages
		To(string) Emitter
		// OnMessage registers a callback which fires when native websocket message received
		OnMessage(NativeMessageFunc)
		// On registers a callback to a particular event which fires when a message to this event received
		On(string, MessageFunc)
		// Join join a connection to a room, it doesn't check if connection is already there, so care
		Join(string)
		// Leave removes a connection from a room
		// Returns true if the connection has actually left from the particular room.
		Leave(string) bool
		// OnLeave registeres a callback which fires when this connection left from any joined room.
		// This callback is called automatically on Disconnected client, because websocket server automatically
		// deletes the disconnected connection from any joined rooms.
		//
		// Note: the callback(s) called right before the server deletes the connection from the room
		// so the connection theoritical can still send messages to its room right before it is being disconnected.
		OnLeave(roomLeaveCb LeaveRoomFunc)
		// Disconnect disconnects the client, close the underline websocket conn and removes it from the conn list
		// returns the error, if any, from the underline connection
		Disconnect() error
		// SetValue sets a key-value pair on the connection's mem store.
		SetValue(key string, value interface{})
		// GetValue gets a value by its key from the connection's mem store.
		GetValue(key string) interface{}
		// GetValueArrString gets a value as []string by its key from the connection's mem store.
		GetValueArrString(key string) []string
		// GetValueString gets a value as string by its key from the connection's mem store.
		GetValueString(key string) string
		// GetValueInt gets a value as integer by its key from the connection's mem store.
		GetValueInt(key string) int
	}

	connection struct {
		underline                UnderlineConnection
		id                       string
		messageType              int
		pinger                   *time.Ticker
		disconnected             bool
		onDisconnectListeners    []DisconnectFunc
		onRoomLeaveListeners     []LeaveRoomFunc
		onErrorListeners         []ErrorFunc
		onNativeMessageListeners []NativeMessageFunc
		onEventListeners         map[string][]MessageFunc
		// these were  maden for performance only
		self      Emitter // pre-defined emitter than sends message to its self client
		broadcast Emitter // pre-defined emitter that sends message to all except this
		all       Emitter // pre-defined emitter which sends message to all clients

		// access to the Context, use with causion, you can't use response writer as you imagine.
		ctx    *iris.Context
		values ConnectionValues
		server *server
		// #119 , websocket writers are not protected by locks inside the gorilla's websocket code
		// so we must protect them otherwise we're getting concurrent connection error on multi writers in the same time.
		writerMu sync.Mutex
		// same exists for reader look here: https://godoc.org/github.com/gorilla/websocket#hdr-Control_Messages
		// but we only use one reader in one goroutine, so we are safe.
		// readerMu sync.Mutex
	}
)

var _ Connection = &connection{}

func newConnection(s *server, ctx *iris.Context, underlineConn UnderlineConnection, id string) *connection {
	c := &connection{
		underline:                underlineConn,
		id:                       id,
		messageType:              websocket.TextMessage,
		onDisconnectListeners:    make([]DisconnectFunc, 0),
		onRoomLeaveListeners:     make([]LeaveRoomFunc, 0),
		onErrorListeners:         make([]ErrorFunc, 0),
		onNativeMessageListeners: make([]NativeMessageFunc, 0),
		onEventListeners:         make(map[string][]MessageFunc, 0),
		ctx:                      ctx,
		server:                   s,
	}

	if s.config.BinaryMessages {
		c.messageType = websocket.BinaryMessage
	}

	c.self = newEmitter(c, c.id)
	c.broadcast = newEmitter(c, Broadcast)
	c.all = newEmitter(c, All)

	return c
}

// write writes a raw websocket message with a specific type to the client
// used by ping messages and any CloseMessage types.
func (c *connection) write(websocketMessageType int, data []byte) {
	// for any-case the app tries to write from different goroutines,
	// we must protect them because they're reporting that as bug...
	c.writerMu.Lock()
	if writeTimeout := c.server.config.WriteTimeout; writeTimeout > 0 {
		// set the write deadline based on the configuration
		c.underline.SetWriteDeadline(time.Now().Add(writeTimeout))
	}

	// .WriteMessage same as NextWriter and close (flush)
	err := c.underline.WriteMessage(websocketMessageType, data)
	c.writerMu.Unlock()
	if err != nil {
		// if failed then the connection is off, fire the disconnect
		c.Disconnect()
	}
}

// writeDefault is the same as write but the message type is the configured by c.messageType
// if BinaryMessages is enabled then it's raw []byte as you expected to work with protobufs
func (c *connection) writeDefault(data []byte) {
	c.write(c.messageType, data)
}

const (
	// WriteWait is 1 second at the internal implementation,
	// same as here but this can be changed at the future*
	WriteWait = 1 * time.Second
)

func (c *connection) startPinger() {

	// this is the default internal handler, we just change the writeWait because of the actions we must do before
	// the server sends the ping-pong.

	pingHandler := func(message string) error {
		err := c.underline.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(WriteWait))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		return err
	}

	c.underline.SetPingHandler(pingHandler)

	// start a new timer ticker based on the configuration
	c.pinger = time.NewTicker(c.server.config.PingPeriod)

	go func() {
		for {
			// wait for each tick
			<-c.pinger.C
			// try to ping the client, if failed then it disconnects
			c.write(websocket.PingMessage, []byte{})
		}
	}()
}

func (c *connection) startReader() {
	conn := c.underline
	hasReadTimeout := c.server.config.ReadTimeout > 0

	conn.SetReadLimit(c.server.config.MaxMessageSize)
	conn.SetPongHandler(func(s string) error {
		if hasReadTimeout {
			conn.SetReadDeadline(time.Now().Add(c.server.config.ReadTimeout))
		}

		return nil
	})

	defer func() {
		c.Disconnect()
	}()

	for {
		if hasReadTimeout {
			// set the read deadline based on the configuration
			conn.SetReadDeadline(time.Now().Add(c.server.config.ReadTimeout))
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				c.EmitError(err.Error())
			}
			break
		} else {
			c.messageReceived(data)
		}

	}

}

// messageReceived checks the incoming message and fire the nativeMessage listeners or the event listeners (ws custom message)
func (c *connection) messageReceived(data []byte) {

	if bytes.HasPrefix(data, websocketMessagePrefixBytes) {
		customData := string(data)
		//it's a custom ws message
		receivedEvt := getWebsocketCustomEvent(customData)
		listeners := c.onEventListeners[receivedEvt]
		if listeners == nil { // if not listeners for this event exit from here
			return
		}
		customMessage, err := websocketMessageDeserialize(receivedEvt, customData)
		if customMessage == nil || err != nil {
			return
		}

		for i := range listeners {
			if fn, ok := listeners[i].(func()); ok { // its a simple func(){} callback
				fn()
			} else if fnString, ok := listeners[i].(func(string)); ok {

				if msgString, is := customMessage.(string); is {
					fnString(msgString)
				} else if msgInt, is := customMessage.(int); is {
					// here if server side waiting for string but client side sent an int, just convert this int to a string
					fnString(strconv.Itoa(msgInt))
				}

			} else if fnInt, ok := listeners[i].(func(int)); ok {
				fnInt(customMessage.(int))
			} else if fnBool, ok := listeners[i].(func(bool)); ok {
				fnBool(customMessage.(bool))
			} else if fnBytes, ok := listeners[i].(func([]byte)); ok {
				fnBytes(customMessage.([]byte))
			} else {
				listeners[i].(func(interface{}))(customMessage)
			}

		}
	} else {
		// it's native websocket message
		for i := range c.onNativeMessageListeners {
			c.onNativeMessageListeners[i](data)
		}
	}

}

func (c *connection) ID() string {
	return c.id
}

func (c *connection) Context() *iris.Context {
	return c.ctx
}

func (c *connection) Values() ConnectionValues {
	return c.values
}

func (c *connection) fireDisconnect() {
	for i := range c.onDisconnectListeners {
		c.onDisconnectListeners[i]()
	}
}

func (c *connection) OnDisconnect(cb DisconnectFunc) {
	c.onDisconnectListeners = append(c.onDisconnectListeners, cb)
}

func (c *connection) OnError(cb ErrorFunc) {
	c.onErrorListeners = append(c.onErrorListeners, cb)
}

func (c *connection) EmitError(errorMessage string) {
	for _, cb := range c.onErrorListeners {
		cb(errorMessage)
	}
}

func (c *connection) To(to string) Emitter {
	if to == Broadcast { // if send to all except me, then return the pre-defined emitter, and so on
		return c.broadcast
	} else if to == All {
		return c.all
	} else if to == c.id {
		return c.self
	}
	// is an emitter to another client/connection
	return newEmitter(c, to)
}

func (c *connection) EmitMessage(nativeMessage []byte) error {
	return c.self.EmitMessage(nativeMessage)
}

func (c *connection) Emit(event string, message interface{}) error {
	return c.self.Emit(event, message)
}

func (c *connection) OnMessage(cb NativeMessageFunc) {
	c.onNativeMessageListeners = append(c.onNativeMessageListeners, cb)
}

func (c *connection) On(event string, cb MessageFunc) {
	if c.onEventListeners[event] == nil {
		c.onEventListeners[event] = make([]MessageFunc, 0)
	}

	c.onEventListeners[event] = append(c.onEventListeners[event], cb)
}

func (c *connection) Join(roomName string) {
	c.server.Join(roomName, c.id)
}

func (c *connection) Leave(roomName string) bool {
	return c.server.Leave(roomName, c.id)
}

func (c *connection) OnLeave(roomLeaveCb LeaveRoomFunc) {
	c.onRoomLeaveListeners = append(c.onRoomLeaveListeners, roomLeaveCb)
	// note: the callbacks are called from the server on the '.leave' and '.LeaveAll' funcs.
}

func (c *connection) fireOnLeave(roomName string) {
	// fire the onRoomLeaveListeners
	for i := range c.onRoomLeaveListeners {
		c.onRoomLeaveListeners[i](roomName)
	}
}

func (c *connection) Disconnect() error {
	return c.server.Disconnect(c.ID())
}

// mem per-conn store

func (c *connection) SetValue(key string, value interface{}) {
	c.values.Set(key, value)
}

func (c *connection) GetValue(key string) interface{} {
	return c.values.Get(key)
}

func (c *connection) GetValueArrString(key string) []string {
	if v := c.values.Get(key); v != nil {
		if arrString, ok := v.([]string); ok {
			return arrString
		}
	}
	return nil
}

func (c *connection) GetValueString(key string) string {
	if v := c.values.Get(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *connection) GetValueInt(key string) int {
	if v := c.values.Get(key); v != nil {
		if i, ok := v.(int); ok {
			return i
		} else if s, ok := v.(string); ok {
			if iv, err := strconv.Atoi(s); err == nil {
				return iv
			}
		}
	}
	return 0
}
