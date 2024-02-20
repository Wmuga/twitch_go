package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/wmuga/twitch_go/internal/bot"
)

// Loads bot options from json file
func Load() (*bot.BotOptions, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path.Join(wd, "configs", "bot_options.json"))
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	opt := &bot.BotOptions{}
	err = json.Unmarshal(bytes, opt)
	if err != nil {
		return nil, err
	}
	return opt, nil
}

func main() {
	opt, err := Load()
	if err != nil {
		fmt.Println(err)
		return
	}
	wg := &sync.WaitGroup{}
	bot.NewBot(opt, wg)
	fmt.Println("Started bot")
	wg.Wait()
}
