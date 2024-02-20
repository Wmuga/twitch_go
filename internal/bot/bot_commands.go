package bot

// TODO: может переделать под геттеры с "middleware?" и проверки на него повесить

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/wmuga/twitch_go/internal/tools"
)

const (
	strNoPermisson = "Не трожь кнопку"
	strWrongArgs   = "Неверный формат команды"

	strSrOwner       = "Включены запросы от стримлера"
	strSrAll         = "Включены запросы для всех"
	strSrIsOff       = "Ничего не включено"
	strSrIsOwnerOnly = "Не для публики"

	strPointsLow    = "Не достаточно поинтов"
	strPointsFormat = "Твое количество поинтов: %d"

	strRollFormat = "%dd%d: %s = %d"

	strRouletteFail          = "Не повезло. Пока, поинты"
	strRouletteSuccessFormat = "Опа. Повезло. Держи %d поинтов"

	strCmds = "Команды. В скобочках скоращенные версии: !help !sr !(p)oints !(ro)ulette !(r)oll"
)

var (
	rollRegex = regexp.MustCompile(`(\d+) *d *(\d+)`)
)

// Handles command with given arguments
func (b *Bot) HandleCommand(channel, sender, command string, args []string, isMod, elfed bool) {
	switch command {
	// Музыка
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
	// Связанное с поинтами
	case "p":
		fallthrough
	case "points":
		b.pointsCommand(channel, sender, elfed)
	case "ro":
		fallthrough
	case "roulette":
		b.rouletteCommand(channel, sender, args, elfed)
	// Остальное
	case "r":
		fallthrough
	case "roll":
		b.rollCommand(channel, sender, strings.Join(args, " "), elfed)
	case "help":
		b.helpCommand(channel, sender, elfed)
	}

}

// Checks permission for using command
func (b *Bot) checkPermission(username string, isMod, ownerOnly bool) bool {
	channelName := "#" + strings.ToLower(username)
	if b.options.Channel == channelName {
		return true
	}

	if !ownerOnly && isMod {
		return true
	}

	return false
}

// Gives help about commands in chat
func (b *Bot) helpCommand(channel, username string, elfed bool) {
	b.replyElfed(channel, username, strCmds, elfed)
}

func (b *Bot) soundCommand() {

}

// Starts music requests. Arg "me|self" set requests as streamer-only
func (b *Bot) startMusicCommand(channel, username, arg string, isMod, elfed bool) {
	if !b.checkPermission(username, isMod, true) {
		b.replyElfed(channel, username, strNoPermisson, elfed)
		return
	}

	b.ytMus.Play()

	if arg == "self" || arg == "me" {
		b.srOwnerOnly = true
		b.sendMessageElfed(channel, strSrOwner, elfed)
		return
	}

	b.srOwnerOnly = false
	b.sendMessageElfed(channel, strSrAll, elfed)
}

// Stops music requests
func (b *Bot) stopMusicCommand(channel, username string, isMod, elfed bool) {
	if !b.checkPermission(username, isMod, true) {
		b.replyElfed(channel, username, strNoPermisson, elfed)
		return
	}

	b.ytMus.Stop()
}

// Add music to music request queue.
func (b *Bot) addMusicCommand(channel, username, data string, isMod, elfed bool) {
	if !b.ytMus.Ready() {
		// if owner - turn on
		if b.checkPermission(username, isMod, true) {
			b.ytMus.Play()
			b.srOwnerOnly = true
		} else {
			// else - nope
			b.replyElfed(channel, username, strSrIsOff, elfed)
			return
		}
	}
	// Check for streamer-only mode
	if b.srOwnerOnly && !b.checkPermission(username, isMod, true) {
		b.replyElfed(channel, username, strSrIsOwnerOnly, elfed)
		return
	}
	// Check if moderator
	if !b.checkPermission(username, isMod, false) && !b.db.TryRemovePoints(username, 5) {
		b.replyElfed(channel, username, strPointsLow, elfed)
		return
	}
	// Try to add music to queue
	msg := b.ytMus.Add(username, isMod, data)
	b.replyElfed(channel, username, msg, elfed)
}

// Skips currently playing music
func (b *Bot) skipMusicCommand(channel, username string, isMod, elfed bool) {
	if !b.checkPermission(username, isMod, true) {
		b.replyElfed(channel, username, strNoPermisson, elfed)
		return
	}

	b.ytMus.Skip()
}

// Gets points of viewer
func (b *Bot) pointsCommand(channel, username string, elfed bool) {
	count := b.db.GetPoints(username)
	b.replyElfed(channel, username, fmt.Sprintf(strPointsFormat, count), elfed)
}

// "Gambling" points
func (b *Bot) rouletteCommand(channel, username string, args []string, elfed bool) {
	if len(args) < 2 {
		b.replyElfed(channel, username, strWrongArgs, elfed)
		return
	}

	points := max(1, tools.NoErrConv(args[0]))
	chance := max(2, tools.NoErrConv(args[1]))

	if !b.db.TryRemovePoints(username, int(points)) {
		b.replyElfed(channel, username, strPointsLow, elfed)
		return
	}

	if rand.Int63n(chance) > 0 {
		b.replyElfed(channel, username, strRouletteFail, elfed)
		return
	}

	b.db.AddPoints(username, int(points*chance))
	b.replyElfed(channel, username,
		fmt.Sprintf(strRouletteSuccessFormat, points*(chance-1)),
		elfed)
}

// Rolls dice. {count}d{sides}
func (b *Bot) rollCommand(channel, username, arg string, elfed bool) {
	data := rollRegex.FindStringSubmatch(arg)
	if len(data) < 3 {
		b.replyElfed(channel, username, strWrongArgs, elfed)
		return
	}

	count := int(max(1, min(tools.NoErrConv(data[1]), 10)))
	size := int(max(2, tools.NoErrConv(data[2])))

	sum := 0
	rolls := make([]string, count)
	for i := 0; i < count; i++ {
		num := rand.Intn(size) + 1
		rolls[i] = strconv.FormatInt(int64(num), 10)
		sum += num
	}
	b.replyElfed(channel, username,
		fmt.Sprintf(strRollFormat, count, size, strings.Join(rolls, " + "), sum),
		elfed)
}

/*
func (B *Bot) changeVolumeCommand() {

}*/
