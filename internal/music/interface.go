package music

type Info struct {
	Album    string `json:"album"`
	Artist   string `json:"artist"`
	Track    string `json:"track"`
	Duration int64  `json:"duration"`
	Username string `json:"username"`
	ID       string `json:"id"`
}

var InfoEmpty = Info{}

type IMusicPlayer interface {
	Play()
	Stop()
	Skip()
	Ready() bool
	Add(username string, isMod bool, search string) string
	ChangeVolume(volume float64)
	Current() Info

	isPlaying() bool
}
