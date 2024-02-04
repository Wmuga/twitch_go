package bot

import "strings"

const (
	strNoPermisson = "Не трожь кнопку"
	strSrOwner     = "Включены запросы от стримлера"
	strSrAll       = "Включены запросы для всех"
)

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

func (b *Bot) helpCommand() {

}

func (b *Bot) soundCommand() {

}

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

func (b *Bot) stopMusicCommand(channel, username string, isMod, elfed bool) {
	if !b.checkPermission(username, isMod, true) {
		b.replyElfed(channel, username, strNoPermisson, elfed)
		return
	}

	b.ytMus.Stop()
}

func (b *Bot) addMusicCommand(channel, username, data string, isMod, elfed bool) {
	if !b.ytMus.Ready() {
		// if owner - turn on
		if b.checkPermission(username, isMod, true) {
			b.ytMus.Play()
		} else {
			// else - nope
			b.replyElfed(channel, username, "Ничего не включено", elfed)
			return
		}
	}

	if b.srOwnerOnly && !b.checkPermission(username, isMod, true) {
		b.replyElfed(channel, username, "Не для публики", elfed)
		return
	}

	if !b.checkPermission(username, isMod, false) && !b.db.TryRemovePoints(username, 5) {
		b.replyElfed(channel, username, "Не достаточно поинтов", elfed)
		return
	}
	msg := b.ytMus.Add(username, isMod, data)
	b.replyElfed(channel, username, msg, elfed)
}

func (b *Bot) skipMusicCommand() {

}

func (b *Bot) pointsCommand() {

}

func (b *Bot) rouletteCommand() {

}

func (b *Bot) rollCommand() {

}

/*
func (B *Bot) changeVolumeCommand() {

}*/
