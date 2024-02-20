package ui

import "github.com/wmuga/twitch_go/internal/music"

// User interface
type UI interface {
	// Sets send event handler
	OnSend(callback SendEventHandler)
	// Sets send-self event handler
	OnSendSelf(callback SendSelfEventHandler)
	// Sets DBUpdate event handler
	OnDBUpdate(callback DBUpdateEventHandler)
	// Sets OnDBGet event handler
	OnDBGet(callback DBGetEventHandler)
	// Sets OnResize event handler
	OnResize(callback ResizeEventHandler)
	// Sets OnCommand event handler
	OnCommand(callback ChatCommandEventHandler)
	// Manual event invoke
	Invoke(event Event, args ...any)
	// Send string to UI
	SendString(str string)
	// Sets current music to UI
	SendMusic(music music.Info)
}
