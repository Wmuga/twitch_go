package music

type Info struct {
	Album    string
	Artist   string
	Track    string
	Duration int64
	Username string
	ID       string
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
