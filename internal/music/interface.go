package music

// Info about music requests
type Info struct {
	Album    string `json:"album"`
	Artist   string `json:"artist"`
	Track    string `json:"track"`
	Duration int64  `json:"duration"`
	Username string `json:"username"`
	ID       string `json:"id"`
}

// Empty Info
var InfoEmpty = Info{}

// Musical player interaction interface
type IMusicPlayer interface {
	// Start playing music from queue
	Play()
	// Stop playing music
	Stop()
	// Skip current playing music
	Skip()
	// Player is ready to play
	Ready() bool
	// Add request to queue
	Add(username string, isMod bool, search string) string
	// Change current volume 0 - 0.1
	ChangeVolume(volume float64)
	// Info about currently playing music
	Current() Info
}
