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
	"golang.org/x/exp/maps"
)

type Bot struct {
	db          database.DBConnection
	conn        *tmi.Connection
	joined      map[string]struct{}
	viewers     map[string]struct{}
	options     *BotOptions
	deelfer     *deelfer.Deelfer
	ytMus       music.IMusicPlayer
	srOwnerOnly bool
}

func New(opt *BotOptions, wg *sync.WaitGroup) *Bot {
	conn := tmi.Connect(opt.Identity.Name, opt.Identity.Oauth)
	b := &Bot{
		conn:    conn,
		options: opt,
		joined:  map[string]struct{}{},
		viewers: map[string]struct{}{},
		deelfer: deelfer.New(),
		db:      database.New(),
		ytMus:   music.New(opt.Youtube.APIKey, opt.Channel[1:]),
	}

	wg.Add(1)
	// messages
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
			b.db.MassAddPoints(maps.Keys(b.viewers), 1)
			time.Sleep(time.Minute)
		}
	}()

	b.Join(opt.Channel)

	fmt.Println("Bot ready")

	return b
}

func (b *Bot) Join(channel string) {
	channel = checkChannel(channel)
	b.joined[channel] = struct{}{}
	b.conn.Join(channel)
}

func (b *Bot) Part(channel string) {
	channel = checkChannel(channel)
	if _, ex := b.joined[channel]; !ex {
		return
	}
	b.Part(channel)
	delete(b.joined, channel)
}

func (b *Bot) SendMessage(channel, message string) {
	channel = checkChannel(channel)
	if _, ex := b.joined[channel]; ex {
		b.Join(channel)
	}
	b.conn.Sendf("PRIVMSG %s :%s", channel, message)
}

func (b *Bot) Reply(channel, dest, message string) {
	b.SendMessage(channel, fmt.Sprintf("@%s, %s", dest, message))
}

func (b *Bot) sendMessageElfed(channel, message string, elfed bool) {
	if elfed {
		b.SendMessage(channel, b.deelfer.Translate(message))
		return
	}
	b.SendMessage(channel, message)
}

func (b *Bot) replyElfed(channel, dest, message string, elfed bool) {
	if elfed {
		b.Reply(channel, dest, b.deelfer.Translate(message))
		return
	}
	b.Reply(channel, dest, message)
}

func (b *Bot) HandleMessage(msg *tmi.Message) {
	if len(msg.Params) == 0 {
		return
	}

	channel := msg.Params[0]
	if channel != b.options.Channel {
		return
	}

	if msg.Command == "JOIN" {
		b.viewers[msg.From] = struct{}{}
		return
	}

	if msg.Command == "PART" {
		delete(b.viewers, msg.From)
		return
	}

	if msg.Command != "PRIVMSG" || !strings.HasPrefix(msg.Trailing, "!") {
		return
	}

	msgStr := strings.Trim(msg.Trailing[1:], " ")
	command := strings.Split(msgStr, " ")
	elfed := false
	// Commands in English. If Russian letter - de elf them
	if slices.Contains(deelfer.ArR, []rune(command[0])[0]) {
		for i, comm := range command {
			command[i] = b.deelfer.Translate(comm)
		}
		elfed = true
	}

	from := msg.From
	if msg.Tags["display-name"] != "" {
		from = msg.Tags["display-name"]
	}

	isMod := false
	if msg.Tags["mod"] != "" {
		isMod = msg.Tags["mod"] != "0"
	}

	b.HandleCommand(channel, from, command[0], command[1:], isMod, elfed)
}

func (b *Bot) HandleCommand(channel, sender, command string, args []string, isMod, elfed bool) {
	switch command {
	case "sr-start":
		b.startMusicCommand(channel, sender, strings.Join(args, " "), isMod, elfed)
	case "sr-end":
		fallthrough
	case "sr-close":
		fallthrough
	case "sr-stop":
		b.stopMusicCommand(channel, sender, isMod, elfed)
	case "sr-skip":
		b.skipMusicCommand(channel, sender, isMod, elfed)
	case "sr":
		b.addMusicCommand(channel, sender, strings.Join(args, " "), isMod, elfed)
	}

}

func checkChannel(channel string) string {
	if strings.HasPrefix(channel, "#") {
		return channel
	}
	return "#" + channel
}
