package overlay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/wmuga/twitch_go/internal/music"
	"github.com/wmuga/twitch_go/internal/ui"
)

type overlay struct {
	*ui.EventHandler
	port   int
	eLog   *log.Logger
	musURL string
	strURL string
}

type strBody struct {
	Str string `json:"str"`
}

func NewOverlayUI(port int) ui.UI {
	ev := ui.NewHandler()
	ovui := &overlay{
		ev,
		port,
		log.New(os.Stdout, "[OV ERR]: ", log.LUTC),
		fmt.Sprintf("http://localhost:%d/api/mus", port),
		fmt.Sprintf("http://localhost:%d/api/str", port),
	}
	ev.UI = ovui

	return ovui
}

func (ov *overlay) SendMusic(music music.Info) {
	data, err := json.Marshal(&music)
	if err != nil {
		ov.eLog.Println(err)
		return
	}
	req, err := http.NewRequest("POST", ov.musURL, bytes.NewReader(data))
	if err != nil {
		ov.eLog.Println(err)
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		ov.eLog.Println(err)
	}
}

func (ov *overlay) SendString(str string) {
	data, err := json.Marshal(&strBody{str})
	if err != nil {
		ov.eLog.Println(err)
		return
	}

	req, err := http.NewRequest("POST", ov.musURL, bytes.NewReader(data))
	if err != nil {
		ov.eLog.Println(err)
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		ov.eLog.Println(err)
	}
}
