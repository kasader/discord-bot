package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kasader/discord-ping/bot"
	"github.com/kasader/discord-ping/config"
)

func main() {
	err := config.ReadConfig()

	if err != nil {
		fmt.Println(err)
		return
	}

	bot.Start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// <-make(chan struct{})
	// return
}
