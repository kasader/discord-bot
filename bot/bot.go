package bot

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/kasader/discord-bot/config"
)

var (
	Token = flag.String("t", "", "Bot authentication token")
	App   = flag.String("a", "", "Application ID")
	Guild = flag.String("g", "", "Guild ID")
	BotID string
)

var session *discordgo.Session

type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "echo",
		Description: "Say something through a bot",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "message",
				Description: "Contents of the message",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "author",
				Description: "Whether to prepend message's author",
				Type:        discordgo.ApplicationCommandOptionBoolean,
			},
		},
	},
	{
		Name:        "about",
		Description: "Show some information about this bot",
		Options:     []*discordgo.ApplicationCommandOption{},
	},
}

func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (om optionMap) {
	om = make(optionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return
}

func interactionAuthor(i *discordgo.Interaction) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
}

func handleEcho(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {
	builder := new(strings.Builder)
	if v, ok := opts["author"]; ok && v.BoolValue() {
		author := interactionAuthor(i.Interaction)
		builder.WriteString("**" + author.String() + "** says: ")
	}
	builder.WriteString(opts["message"].StringValue())

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
		},
	})

	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}

func handleAbout(s *discordgo.Session, i *discordgo.InteractionCreate) {

	b, err := os.ReadFile("README.md")
	if err != nil {
		fmt.Print(err)
	}

	str := string(b)

	// builder := new(strings.Builder)
	// if v, ok := opts["author"]; ok && v.BoolValue() {
	// 	author := interactionAuthor(i.Interaction)
	// 	builder.WriteString("**" + author.String() + "** says: ")
	// }
	// builder.WriteString(opts["message"].StringValue())

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: str,
		},
	})

	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()

	switch data.Name {
	case "echo":
		handleEcho(s, i, parseOptions(data.Options))
	case "about":
		handleAbout(s, i)
	default:
		return
	}
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

func init() { flag.Parse() }
func Start() {

	// Create a new Discord session with bot token.
	var err error
	session, err := discordgo.New("Bot " + *Token)
	if err != nil {
		log.Fatalf("Error creating Discord session, %v", err)
	}

	// Set BotID variable
	u, err := session.User("@me")
	if err != nil {
		log.Fatal(err)
	}
	BotID = u.ID

	// Register the messageHandler function as a callback for messageCreate events.
	session.AddHandler(messageHandler)
	session.AddHandler(interactionHandler)
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s\n", r.User.String())
	})

	_, err = session.ApplicationCommandBulkOverwrite(*App, *Guild, commands)
	if err != nil {
		log.Fatalf("Could not register commands: %s", err)
	}

	// Listen for all intents.
	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	// Open a webstocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		log.Fatalf("Error opening connection, %v", err)
	}

	fmt.Println("The bot is now running. Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("Shutdown signal received... Exiting.")
	session.Close()
}
