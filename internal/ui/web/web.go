package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/wmuga/twitch_go/internal/music"
	"github.com/wmuga/twitch_go/internal/tools"
	"github.com/wmuga/twitch_go/internal/ui"

	"github.com/gorilla/mux"
)

type WebUI struct {
	*ui.EventHandler
	eLog        *log.Logger
	router      *mux.Router
	connections map[int]chan webEvent
	conMux      *sync.RWMutex
	last        int
}

type webEvent struct {
	Event string
	Data  string
}

func New(port int) ui.UI {
	ev := ui.NewHandler()
	r := mux.NewRouter()

	web := &WebUI{
		ev,
		log.New(os.Stdout, "[WEB ERR]: ", log.LUTC),
		r,
		make(map[int]chan webEvent),
		&sync.RWMutex{},
		0,
	}
	ev.UI = web

	r.PathPrefix("/index").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui.html", http.StatusPermanentRedirect)
	})
	r.HandleFunc("/sse", web.sse)
	r.HandleFunc("/send", web.sendHandler)
	r.HandleFunc("/sendself", web.sendSelfHandler)
	r.HandleFunc("/dbget", web.dbGetHandler)
	r.HandleFunc("/dbupdate", web.dbUpdateHandler)
	r.HandleFunc("/resize", web.resizeHandler)
	r.HandleFunc("/command", web.commandHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	go func() {
		fmt.Println(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), r))
	}()

	fmt.Println("Web UI working on port", port)

	return web
}

func (web *WebUI) SendMusic(music music.Info) {
	bytes, err := json.Marshal(&music)
	if err != nil {
		web.eLog.Println(err)
		return
	}

	web.sendEvent(webEvent{
		Event: "music",
		Data:  string(bytes),
	})
}

func (web *WebUI) SendString(str string) {
	web.sendEvent(webEvent{
		Event: "str",
		Data:  str,
	})
}

func (web *WebUI) sendEvent(ev webEvent) {
	web.conMux.RLock()
	for _, c := range web.connections {
		c <- ev
	}
	web.conMux.RUnlock()
}

func (web *WebUI) sse(w http.ResponseWriter, r *http.Request) {
	// new connection id
	id := web.last
	web.last++
	// add new channel to current connections
	infoChan := make(chan webEvent)
	web.conMux.Lock()
	web.connections[id] = infoChan
	web.conMux.Unlock()
	// remove on exit
	defer func() {
		web.conMux.Lock()
		delete(web.connections, id)
		web.conMux.Unlock()
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-infoChan:
			msg := fmt.Sprintf("%s: %s\n\n", event.Event, event.Data)
			_, err := w.Write([]byte(msg))
			if err != nil {
				web.eLog.Println(err)
			}
		}
	}
}

func (web *WebUI) sendHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ch := q.Get("chan")
	msg := q.Get("msg")
	web.Invoke(ui.Send, ch, msg)
}

func (web *WebUI) sendSelfHandler(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	web.Invoke(ui.SendSelf, msg)
}

func (web *WebUI) dbGetHandler(w http.ResponseWriter, r *http.Request) {
	web.Invoke(ui.DBGet)
}

func (web *WebUI) dbUpdateHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	usr := q.Get("usr")
	pts := q.Get("pts")
	web.Invoke(ui.DBUpdate, usr, tools.NoErrConv(pts))
}

func (web *WebUI) resizeHandler(w http.ResponseWriter, r *http.Request) {
	big := r.URL.Query().Get("big")
	web.Invoke(ui.Resize, big == "true")
}

func (web *WebUI) commandHandler(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Query().Get("cmd")
	web.Invoke(ui.Command, cmd)
}
