package ui

import "github.com/wmuga/twitch_go/internal/music"

type UI interface {
	OnSend(callback SendEventHandler)
	OnSendSelf(callback SendSelfEventHandler)
	OnDBUpdate(callback DBUpdateEventHandler)
	OnDBGet(callback DBGetEventHandler)
	OnResize(callback ResizeEventHandler)
	OnCommand(callback ChatCommandEventHandler)
	Invoke(event Event, args ...any)
	SendString(str string)
	SendMusic(music music.Info)
}
