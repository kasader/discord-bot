package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kasader/discord-ping/config"
)

var BotID string
var sess *discordgo.Session

func Start() {
	sess, err := discordgo.New("Bot " + config.Token)

	if err != nil {
		log.Fatal(err) // log.Fatal() is a call to log(err), and then os.exit()
	}

	u, err := sess.User("@me")

	if err != nil {
		log.Fatal(err)
	}

	BotID = u.ID

	sess.AddHandler(messageHandler)
	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()

	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("The bot is online!")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Make sure that the bot is NOT responding to its own messages
	if m.Author.ID == BotID {
		return
	}

	args := strings.Split(m.Content, " ")
	if args[0] == config.BotPrefix+"echo" {
		s.ChannelMessageSend(m.ChannelID, strings.TrimPrefix(m.Content, config.BotPrefix+"echo"))
	}

	if m.Content == config.BotPrefix+"ping" {
		s.ChannelMessageSend(m.ChannelID, "pong")
	}

	if m.Content == config.BotPrefix+"hello" {
		s.ChannelMessageSend(m.ChannelID, "world!")
	}

	author := discordgo.MessageEmbedAuthor{
		Name: "John Pork",
		URL:  "https://google.com",
	}
	embed := discordgo.MessageEmbed{
		Title:  "words",
		Author: &author,
	}

	if m.Content == config.BotPrefix+"embed" {
		s.ChannelMessageSendEmbed(m.ChannelID, &embed)
	}
}
