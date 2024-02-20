package bot

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/sunspots/tmi"
	"github.com/wmuga/twitch_go/internal/database"
	"github.com/wmuga/twitch_go/internal/deelfer"
	"github.com/wmuga/twitch_go/internal/music"
	"github.com/wmuga/twitch_go/internal/ui"
	"github.com/wmuga/twitch_go/internal/ui/overlay"
	"github.com/wmuga/twitch_go/internal/ui/web"
	"golang.org/x/exp/maps"
)

// Main bot struct.
// Use NewBot to create new instance.
type Bot struct {
	db          database.DBConnection
	conn        *tmi.Connection
	joined      map[string]struct{}
	curViewers  map[string]struct{}
	viewers     map[string]struct{}
	options     *BotOptions
	deelfer     *deelfer.Deelfer
	ytMus       music.IMusicPlayer
	srOwnerOnly bool
	overlay     ui.UI
	uis         []ui.UI
}

// Creates new instance of Bot.
func NewBot(opt *BotOptions, wg *sync.WaitGroup) *Bot {
	conn := tmi.Connect(opt.Identity.Name, opt.Identity.Oauth)
	b := &Bot{
		conn:       conn,
		options:    opt,
		joined:     map[string]struct{}{},
		curViewers: map[string]struct{}{},
		viewers:    map[string]struct{}{},
		deelfer:    deelfer.NewDeelfer(),
		db:         database.NewSqlite(),
		ytMus:      music.NewYTMusic(opt.Youtube.APIKey, opt.Channel[1:]),
		uis:        []ui.UI{web.NewWebUI(opt.UIPort), overlay.NewOverlayUI(opt.OverlayPort)},
	}
	b.overlay = b.uis[1]
	// messages
	wg.Add(1)
	go func() {
		for {
			msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Error:", err)
				break
			}
			b.HandleMessage(msg)
		}
		wg.Done()
	}()
	// add points
	go func() {
		for {
			b.db.MassAddPoints(maps.Keys(b.curViewers), 1)
			time.Sleep(time.Minute)
		}
	}()

	b.setupUIs()
	b.updateSong()

	b.Join(opt.Channel)

	fmt.Println("Bot ready")

	return b
}

// Joins twitch channel's chat
func (b *Bot) Join(channel string) {
	channel = checkChannel(channel)
	b.joined[channel] = struct{}{}
	b.conn.Join(channel)
}

// Leaves twitch channel's chat
func (b *Bot) Part(channel string) {
	channel = checkChannel(channel)
	if _, ex := b.joined[channel]; !ex {
		return
	}
	b.Part(channel)
	delete(b.joined, channel)
}

// Send message to twitch channel's chat
func (b *Bot) SendMessage(channel, message string) {
	channel = checkChannel(channel)
	if _, ex := b.joined[channel]; ex {
		b.Join(channel)
	}
	b.conn.Sendf("PRIVMSG %s :%s", channel, message)
}

// Replies to viewer in twitch chat
func (b *Bot) Reply(channel, dest, message string) {
	b.SendMessage(channel, fmt.Sprintf("@%s, %s", dest, message))
}

// Sends message to twitch channel's chat with optional "Elfed" effect
func (b *Bot) sendMessageElfed(channel, message string, elfed bool) {
	if elfed {
		b.SendMessage(channel, b.deelfer.Translate(message))
		return
	}
	b.SendMessage(channel, message)
}

// Replies to viewer in twitch chat with optional "Elfed" effect
func (b *Bot) replyElfed(channel, dest, message string, elfed bool) {
	if elfed {
		b.Reply(channel, dest, b.deelfer.Translate(message))
		return
	}
	b.Reply(channel, dest, message)
}

// Handles messages from joined channels
func (b *Bot) HandleMessage(msg *tmi.Message) {
	// skips own messages
	if msg.From == b.options.Identity.Name {
		return
	}
	// Skips PING messages
	if len(msg.Params) == 0 {
		return
	}
	// Skips non moderated channels
	channel := msg.Params[0]
	if channel != b.options.Channel {
		return
	}
	// Greets new viewers
	if msg.Command == "JOIN" {
		b.curViewers[msg.From] = struct{}{}
		if _, ex := b.viewers[msg.From]; ex || msg.From == channel[1:] {
			return
		}
		b.viewers[msg.From] = struct{}{}

		hello := fmt.Sprintf("Привествую, @%s", msg.From)
		b.overlay.SendString(hello)
		b.SendMessage(channel, hello)
		return
	}
	// Removes leaving viewers
	if msg.Command == "PART" {
		delete(b.curViewers, msg.From)
		return
	}
	// Skips other messages without '!'
	if msg.Command != "PRIVMSG" || !strings.HasPrefix(msg.Trailing, "!") {
		return
	}

	msgStr := strings.Trim(msg.Trailing[1:], " ")
	command := strings.Split(msgStr, " ")
	elfed := false
	// Commands in English. If Russian letter - de "elf" them
	if slices.Contains(deelfer.ArR, []rune(command[0])[0]) {
		for i, comm := range command {
			command[i] = b.deelfer.Translate(comm)
		}
		elfed = true
	}
	// Extract Display name
	from := msg.From
	if msg.Tags["display-name"] != "" {
		from = msg.Tags["display-name"]
	}
	// Extract isMod
	isMod := false
	if msg.Tags["mod"] != "" {
		isMod = msg.Tags["mod"] != "0"
	}
	// Handle command
	b.HandleCommand(channel, from, command[0], command[1:], isMod, elfed)
}

// Set event handlers for uis
func (b *Bot) setupUIs() {
	owner := b.options.Channel[1:]
	for _, ui := range b.uis {
		ui.OnSend(func(channel, message string) {
			b.SendMessage(channel, message)
		})

		ui.OnSendSelf(func(message string) {
			b.SendMessage(b.options.Channel, message)
		})

		ui.OnCommand(func(cmd string) {
			data := strings.Split(cmd, " ")
			b.HandleCommand(b.options.Channel, owner, data[0][1:], data[1:], false, false)
		})

		ui.OnDBGet(func() {
			ui.SendString("Not implemented")
		})

		ui.OnDBUpdate(func(usr string, pts int64) {
			ui.SendString("Not implemented")
		})

		ui.OnResize(func(big bool) {
			ui.SendString("Not implemented")
		})
	}
}

// Sends music to UIs if changed
func (b *Bot) updateSong() {
	last := music.InfoEmpty
	go func() {
		for {
			cur := b.ytMus.Current()
			if cur == last {
				continue
			}
			last = cur
			for _, ui := range b.uis {
				ui.SendMusic(last)
			}
			time.Sleep(time.Second)
		}
	}()
}

// Checks if channle has '#' prefix
func checkChannel(channel string) string {
	if strings.HasPrefix(channel, "#") {
		return channel
	}
	return "#" + channel
}
