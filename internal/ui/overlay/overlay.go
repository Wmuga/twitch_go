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
	str string
}

func NewOverlayUI(port int) ui.UI {
	ev := ui.NewHandler()
	ovui := &overlay{
		ev,
		port,
		log.New(os.Stdout, "[OV ERR]: ", log.LUTC),
		fmt.Sprintf("http://localhost:%d/mus", port),
		fmt.Sprintf("http://localhost:%d/str", port),
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

	_, err = http.Post(ov.musURL, "application/json", bytes.NewReader(data))
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

	_, err = http.Post(ov.musURL, "application/json", bytes.NewReader(data))
	if err != nil {
		ov.eLog.Println(err)
	}
}
