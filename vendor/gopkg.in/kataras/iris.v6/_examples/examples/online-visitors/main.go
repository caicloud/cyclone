package main

import (
	"sync/atomic"

	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/view"
	"gopkg.in/kataras/iris.v6/adaptors/websocket"
)

var (
	app *iris.Framework
	ws  websocket.Server
)

func init() {
	// init the server instance
	app = iris.New()
	// adapt a logger in dev mode
	app.Adapt(iris.DevLogger())
	// adapt router
	app.Adapt(httprouter.New())
	// adapt templaes
	app.Adapt(view.HTML("./templates", ".html").Reload(true))
	// adapt websocket
	ws = websocket.New(websocket.Config{Endpoint: "/my_endpoint"})
	ws.OnConnection(HandleWebsocketConnection)
	app.Adapt(ws)
}

type page struct {
	PageID string
}

func main() {
	app.StaticWeb("/js", "./static/assets/js")

	h := func(ctx *iris.Context) {
		ctx.Render("index.html", page{PageID: "index page"})
	}

	h2 := func(ctx *iris.Context) {
		ctx.Render("other.html", page{PageID: "other page"})
	}

	// Open some browser tabs/or windows
	// and navigate to
	// http://localhost:8080/ and http://localhost:8080/other
	// Each page has its own online-visitors counter.
	app.Get("/", h)
	app.Get("/other", h2)
	app.Listen(":8080")
}

type pageView struct {
	source string
	count  uint64
}

func (v *pageView) increment() {
	atomic.AddUint64(&v.count, 1)
}

func (v *pageView) decrement() {
	oldCount := v.count
	if oldCount > 0 {
		atomic.StoreUint64(&v.count, oldCount-1)
	}
}

func (v *pageView) getCount() uint64 {
	val := atomic.LoadUint64(&v.count)
	return val
}

type (
	pageViews []pageView
)

func (v *pageViews) Add(source string) {
	args := *v
	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if kv.source == source {
			kv.increment()
			return
		}
	}

	c := cap(args)
	if c > n {
		args = args[:n+1]
		kv := &args[n]
		kv.source = source
		kv.count = 1
		*v = args
		return
	}

	kv := pageView{}
	kv.source = source
	kv.count = 1
	*v = append(args, kv)
}

func (v *pageViews) Get(source string) *pageView {
	args := *v
	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if kv.source == source {
			return kv
		}
	}
	return nil
}

func (v *pageViews) Reset() {
	*v = (*v)[:0]
}

var v pageViews

// HandleWebsocketConnection handles the online viewers per example(gist source)
func HandleWebsocketConnection(c websocket.Connection) {

	c.On("watch", func(pageSource string) {
		v.Add(pageSource)
		// join the socket to a room linked with the page source
		c.Join(pageSource)

		viewsCount := v.Get(pageSource).getCount()
		if viewsCount == 0 {
			viewsCount++ // count should be always > 0 here
		}
		c.To(pageSource).Emit("watch", viewsCount)
	})

	c.OnLeave(func(roomName string) {
		if roomName != c.ID() { // if the roomName  it's not the connection iself
			// the roomName here is the source, this is the only room(except the connection's ID room) which we join the users to.
			pageV := v.Get(roomName)
			if pageV == nil {
				return // for any case that this room is not a pageView source
			}
			// decrement -1 the specific counter for this page source.
			pageV.decrement()
			// 1. open 30 tabs.
			// 2. close the browser.
			// 3. re-open the browser
			// 4. should be  v.getCount() = 1
			// in order to achieve the previous flow we should decrement exactly when the user disconnects
			// but emit the result a little after, on a goroutine
			// getting all connections within this room and emit the online views one by one.
			// note:
			// we can also add a time.Sleep(2-3 seconds) inside the goroutine at the future if we don't need 'real-time' updates.
			go func(currentConnID string) {
				for _, conn := range ws.GetConnectionsByRoom(roomName) {
					if conn.ID() != currentConnID {
						conn.Emit("watch", pageV.getCount())
					}

				}
			}(c.ID())
		}

	})
}
