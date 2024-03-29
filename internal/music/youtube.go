package music

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	ytdl "github.com/wader/goutubedl"
	at "github.com/wmuga/twitch_go/internal/atomic"
	"github.com/wmuga/twitch_go/internal/tools"
)

// Plays music from youtube.
// Create new instance with NewYTMusic
type YTMusic struct {
	apiKey   string
	owner    string
	durOwner int64
	durMod   int64
	durOther int64
	eLogger  *log.Logger
	current  *at.Type[Info]

	queue    *at.List[Info]
	loopPlay atomic.Bool
	canPlay  chan struct{}

	curStream beep.StreamSeekCloser
}

// Nesessary fields from yt request api
type idResult struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title   string `json:"title"`
			Channel string `json:"channelTitle"`
		} `json:"snippet"`
		Details struct {
			Duration string `json:"duration"`
		} `json:"contentDetails"`
	} `json:"items"`
}

// Nesessary fields from yt request api
type searchResult struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
}

const (
	urlSearchFormat    = "https://www.googleapis.com/youtube/v3/search?part=snippet&maxResults=1&q=%s&type=video&key=%s"
	urlVideoDataFormat = "https://www.googleapis.com/youtube/v3/videos?id=%s&key=%s&part=snippet&part=contentDetails"
	urlVideoFormat     = "https://www.youtube.com/watch?v=%s"

	strTooLong     = "Слишком длинное видео"
	strNoInfo      = "Не удалось найти видео"
	strQueueMax    = "Превышен лимит заказов"
	strAddedFormat = "Видео %s - %s добавлено в очередь: %d"
)

var (
	ytReg1 = regexp.MustCompile(`watch\?v=`)
	ytReg2 = regexp.MustCompile(`tu\.be`)
)

// Creates new instance of YTMusic
func NewYTMusic(apiKey, owner string) IMusicPlayer {
	yt := &YTMusic{
		apiKey:   apiKey,
		owner:    owner,
		durOwner: 30 * 60,
		durMod:   10 * 60,
		durOther: 5 * 60,
		eLogger:  log.New(os.Stdout, "[YT ERR] ", log.LUTC),
		queue:    at.NewList[Info](),
		canPlay:  make(chan struct{}, 1),
		loopPlay: atomic.Bool{},
		current:  at.NewType[Info](),
	}

	yt.loopPlay.Store(false)
	yt.canPlay <- struct{}{}

	go func() {
		for {
			// check if playing is on
			if !yt.loopPlay.Load() {
				time.Sleep(time.Second * 5)
				continue
			}

			select {
			// check if already playing
			case <-yt.canPlay:
				yt.tryPlaying()
			default:
				time.Sleep(time.Second * 5)
			}
		}
	}()

	fmt.Println("YT ready")

	return yt
}

// Implementation of IMusicPlayer.AddMassAddPoints
func (yt *YTMusic) Add(username string, isMod bool, search string) string {
	if !isMod && !yt.isOwner(username) && yt.countRequests(username) > 2 {
		return strQueueMax
	}

	var info Info

	if ytReg1.MatchString(search) {
		id := strings.Split(strings.Split(search, " ")[0], "v=")[1]
		id = strings.Split(id, "&")[0]
		info = yt.getVideoInfo(id)
	} else if ytReg2.MatchString(search) {
		data := strings.Split(strings.Split(search, " ")[0], "be/")
		id := data[len(data)-1]
		info = yt.getVideoInfo(id)
	} else {
		info = yt.search(username, search)
	}

	if info == InfoEmpty {
		return strNoInfo
	}

	if yt.isOwner(username) && info.Duration <= yt.durOwner ||
		isMod && info.Duration <= yt.durMod ||
		info.Duration <= yt.durOther {

		yt.queue.Push(info)
		return fmt.Sprintf(strAddedFormat, info.Artist, info.Track, yt.queue.Count())
	}

	return strTooLong
}

// Implementation of IMusicPlayer.ChangeVolume
func (*YTMusic) ChangeVolume(volume float64) {
	panic("unimplemented")
}

// Implementation of IMusicPlayer.Current
func (yt *YTMusic) Current() Info {
	return yt.current.Load()
}

// Implementation of IMusicPlayer.Ready
func (yt *YTMusic) Ready() bool {
	return yt.loopPlay.Load()
}

// Implementation of IMusicPlayer.Play
func (yt *YTMusic) Play() {
	yt.loopPlay.Store(true)
}

// Implementation of IMusicPlayer.Skip
func (yt *YTMusic) Skip() {
	if !yt.isPlaying() {
		return
	}
	speaker.Close()
	speaker.Clear()
	yt.callback()
}

// Implementation of IMusicPlayer.Stop
func (yt *YTMusic) Stop() {
	yt.loopPlay.Store(false)
	yt.Skip()
}

// Returns if music is playing
func (yt *YTMusic) isPlaying() bool {
	return yt.Ready() && yt.current.Load() != InfoEmpty
}

// Get length of video in seconds
func (*YTMusic) getLength(lenStr string) int64 {
	if len(lenStr) < 2 {
		return 0
	}

	var length int64
	lenStr = lenStr[2:]
	// hours
	data := strings.Split(lenStr, "H")
	if len(data) == 2 {
		length += tools.NoErrConv(data[0]) * 3600
	}

	// minutes
	data = strings.Split(data[len(data)-1], "M")
	if len(data) == 2 {
		length += tools.NoErrConv(data[0]) * 60
	}

	data = strings.Split(data[len(data)-1], "S")
	length += tools.NoErrConv(data[0])

	return length
}

// Returns if requester is owner
func (yt *YTMusic) isOwner(username string) bool {
	return strings.EqualFold(username, yt.owner)
}

// Counts requests of user
func (yt *YTMusic) countRequests(username string) int {
	res := 0
	for _, r := range yt.queue.Elements() {
		if strings.EqualFold(r.Username, username) {
			res++
		}
	}
	return res
}

// Searchs video with given data
func (yt *YTMusic) search(username, data string) Info {
	info := yt.requestYT(urlSearchFormat, url.QueryEscape(data), yt.apiKey)
	if info != InfoEmpty {
		info.Username = username
	}
	return info
}

// Sends request to yt api
func (yt *YTMusic) requestYT(reqUrl string, params ...any) Info {
	reqUrl = fmt.Sprintf(reqUrl, params...)
	resp, err := http.Get(reqUrl)
	if err != nil {
		yt.eLogger.Println("Can't GET:", err)
		return InfoEmpty
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		yt.eLogger.Println("Can't read:", err)
		return InfoEmpty
	}

	search := searchResult{}
	err = json.Unmarshal(data, &search)
	if err != nil {
		yt.eLogger.Println("Can't unmarshal:", err)
		return InfoEmpty
	}

	if len(search.Items) == 0 {
		fmt.Println("No info")
		return InfoEmpty
	}

	return yt.getVideoInfo(search.Items[0].ID.VideoID)
}

// Gets info from video
func (yt *YTMusic) getVideoInfo(id string) Info {
	reqUrl := fmt.Sprintf(urlVideoDataFormat, id, yt.apiKey)
	resp, err := http.Get(reqUrl)
	if err != nil {
		yt.eLogger.Println("Can't GET:", err)
		return InfoEmpty
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		yt.eLogger.Println("Can't read:", err)
		return InfoEmpty
	}

	vids := idResult{}
	err = json.Unmarshal(data, &vids)
	if err != nil {
		yt.eLogger.Println("Can't unmarshal:", err)
		return InfoEmpty
	}

	if len(vids.Items) == 0 {
		return InfoEmpty
	}

	vid := vids.Items[0]

	return Info{
		ID:       vid.ID,
		Artist:   vid.Snippet.Channel,
		Track:    vid.Snippet.Title,
		Duration: yt.getLength(vid.Details.Duration),
	}
}

// tries to start playing video from queue
func (yt *YTMusic) tryPlaying() {
	// check queue length
	if yt.queue.Count() == 0 {
		time.Sleep(time.Second * 5)
		yt.canPlay <- struct{}{}
		return
	}
	// get item from queue
	item, err := yt.queue.Shift()
	if err != nil {
		yt.eLogger.Println(err)
		yt.canPlay <- struct{}{}
		yt.current.Store(InfoEmpty)
		return
	}
	// download video
	buf, err := downloadMusic(item)
	if err != nil {
		yt.eLogger.Println(err)
		yt.canPlay <- struct{}{}
		yt.current.Store(InfoEmpty)
		return
	}
	// convert to mp3 format
	resultBuffer, err := convertToMp3(buf)
	if err != nil {
		yt.eLogger.Println(err)
		yt.canPlay <- struct{}{}
		yt.current.Store(InfoEmpty)
		return
	}
	// Decodes mp3
	reader := bytes.NewReader(resultBuffer.Bytes())
	stream, format, err := mp3.Decode(io.NopCloser(reader))
	if err != nil {
		yt.eLogger.Println(err)
		yt.canPlay <- struct{}{}
		yt.current.Store(InfoEmpty)
		return
	}
	// Inits spreaker
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		yt.eLogger.Println(err)
		yt.canPlay <- struct{}{}
		yt.current.Store(InfoEmpty)
		speaker.Close()
		return
	}
	// Starts playing music
	yt.curStream = stream
	yt.current.Store(item)

	speaker.Play(beep.Seq(stream,
		beep.Callback(yt.callback)))
}

// Downloads video from youtube
func downloadMusic(item Info) ([]byte, error) {
	fmt.Println("downloading")
	res, err := ytdl.New(context.Background(), fmt.Sprintf(urlVideoFormat, item.ID), ytdl.Options{})
	if err != nil {
		return nil, err
	}

	mus, err := res.Download(context.Background(), "bestaudio")
	if err != nil {
		return nil, err
	}

	defer mus.Close()
	buf, err := io.ReadAll(mus)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Uses video to mp3 using ffmpeg
func convertToMp3(buf []byte) (*bytes.Buffer, error) {
	cmd := exec.Command("ffmpeg", "-y", // Yes to all
		"-hide_banner", "-loglevel", "panic", // Hide all logs
		"-i", "pipe:0", // take stdin as input
		"-map_metadata", "-1", // strip out all (mostly) metadata
		"-c:a", "libmp3lame", // use mp3 lame codec
		"-vsync", "2", // suppress "Frame rate very high for a muxer not efficiently supporting it"
		"-b:a", "128k", // Down sample audio birate to 128k
		"-f", "mp3", // using mp3 muxer (IMPORTANT, output data to pipe require manual muxer selecting)
		"pipe:1", // output to stdout
	)
	resultBuffer := bytes.NewBuffer(make([]byte, 60*1024*1024))
	// Binding 0-2 ports
	cmd.Stderr = os.Stderr
	cmd.Stdout = resultBuffer
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	fmt.Println("ffmpeg convertion start")

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	_, err = stdin.Write(buf)
	if err != nil {
		return nil, err
	}

	err = stdin.Close()
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	return resultBuffer, nil
}

// callback for end music event
func (yt *YTMusic) callback() {
	yt.canPlay <- struct{}{}
	yt.current.Store(InfoEmpty)
	yt.curStream.Close() // nolint:errcheck
}
