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

type EventHandler struct {
	UI
	sendHandler     SendEventHandler
	sendSelfHandler SendSelfEventHandler
	dbUpdateHandler DBUpdateEventHandler
	dbGetHandler    DBGetEventHandler
	resizeHandler   ResizeEventHandler
	commandHandler  ChatCommandEventHandler
}

func NewHandler() *EventHandler {
	return &EventHandler{}
}

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

func (ev *EventHandler) OnCommand(callback ChatCommandEventHandler) {
	ev.commandHandler = callback
}

func (ev *EventHandler) OnDBGet(callback DBGetEventHandler) {
	ev.dbGetHandler = callback
}

func (ev *EventHandler) OnDBUpdate(callback DBUpdateEventHandler) {
	ev.dbUpdateHandler = callback
}

func (ev *EventHandler) OnResize(callback ResizeEventHandler) {
	ev.resizeHandler = callback
}

func (ev *EventHandler) OnSend(callback SendEventHandler) {
	ev.sendHandler = callback
}

func (ev *EventHandler) OnSendSelf(callback SendSelfEventHandler) {
	ev.sendSelfHandler = callback
}
