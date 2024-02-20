package ui

type Event string

const (
	Send     Event = "send"
	SendSelf Event = "send-self"
	DBUpdate Event = "db-update"
	DBGet    Event = "db-get"
	Resize   Event = "resize"
	Command  Event = "command"
)

type (
	SendEventHandler        func(channel, message string)
	SendSelfEventHandler    func(message string)
	DBUpdateEventHandler    func(username string, points int64)
	DBGetEventHandler       func()
	ResizeEventHandler      func(small bool)
	ChatCommandEventHandler func(command string)
)

// Event handler implimentation. Invokes callbacks on events.
// Create new instance with NewHandler.
type EventHandler struct {
	UI
	sendHandler     SendEventHandler
	sendSelfHandler SendSelfEventHandler
	dbUpdateHandler DBUpdateEventHandler
	dbGetHandler    DBGetEventHandler
	resizeHandler   ResizeEventHandler
	commandHandler  ChatCommandEventHandler
}

// Creates new instance of EventHandler
func NewHandler() *EventHandler {
	return &EventHandler{}
}

// Implementation of UI.Invoke
func (ev *EventHandler) Invoke(event Event, args ...any) {
	switch event {
	case Send:
		if ev.sendHandler != nil {
			ev.sendHandler(args[0].(string), args[1].(string))
		}
	case SendSelf:
		if ev.sendSelfHandler != nil {
			ev.sendSelfHandler(args[0].(string))
		}
	case DBUpdate:
		if ev.dbUpdateHandler != nil {
			ev.dbUpdateHandler(args[0].(string), args[1].(int64))
		}
	case DBGet:
		if ev.dbGetHandler != nil {
			ev.dbGetHandler()
		}
	case Resize:
		if ev.resizeHandler != nil {
			ev.resizeHandler(args[0].(bool))
		}
	case Command:
		if ev.commandHandler != nil {
			ev.commandHandler(args[0].(string))
		}
	}
}

// Implementation of UI.OnCommand
func (ev *EventHandler) OnCommand(callback ChatCommandEventHandler) {
	ev.commandHandler = callback
}

// Implementation of UI.OnDBGet
func (ev *EventHandler) OnDBGet(callback DBGetEventHandler) {
	ev.dbGetHandler = callback
}

// Implementation of UI.OnDBUpdate
func (ev *EventHandler) OnDBUpdate(callback DBUpdateEventHandler) {
	ev.dbUpdateHandler = callback
}

// Implementation of UI.OnResize
func (ev *EventHandler) OnResize(callback ResizeEventHandler) {
	ev.resizeHandler = callback
}

// Implementation of UI.OnSend
func (ev *EventHandler) OnSend(callback SendEventHandler) {
	ev.sendHandler = callback
}

// Implementation of UI.OnSendSelf
func (ev *EventHandler) OnSendSelf(callback SendSelfEventHandler) {
	ev.sendSelfHandler = callback
}
