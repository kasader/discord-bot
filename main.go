package main

import (
	"fmt"

	"github.com/kasader/discord-bot/bot"
	"github.com/kasader/discord-bot/config"
)

func main() {
	err := config.ReadConfig()

	if err != nil {
		fmt.Println(err)
		return
	}

	bot.Start()

	// Cleanly close down the Discord session.

	// <-make(chan struct{})
	// return
}
