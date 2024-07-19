package bot

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kasader/discord-bot/config"
	//"github.com/mmcdole/gofeed"
)

var (
	Token       = flag.String("t", "", "Bot authentication token")
	App         = flag.String("a", "", "Application ID")
	Guild       = flag.String("g", "", "Guild ID")
	BotID       string
	PathToFeeds = "feeds.json"
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

	{
		Name:        "rss_add_feed",
		Description: "Add an rss feed to the bot",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "feed",
				Description: "URL of the feed to add to the bot",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	},

	{
		Name:        "rss_remove_feed",
		Description: "Remove an RSS feed from the bot",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "feed",
				Description: "URL of the feed to remove from the bot",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	},

	{
		Name:        "rss_list_show",
		Description: "Show the list of currently registered RSS feeds",
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

func handleRssListShow(s *discordgo.Session, i *discordgo.InteractionCreate) {

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00, // Green
		Description: "This is a discordgo embed",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "I am a field",
				Value:  "I am a value",
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "I am a second field",
				Value:  "I am a value",
				Inline: true,
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: "https://cdn.discordapp.com/avatars/119249192806776836/cc32c5c3ee602e1fe252f9f595f9010e.jpg?size=2048",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://cdn.discordapp.com/avatars/119249192806776836/cc32c5c3ee602e1fe252f9f595f9010e.jpg?size=2048",
		},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     "I am an Embed",
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

}

func handleRssRemoveFeed(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {

	//does rss feed already exist?
	var feeds []string

	//is the rss feed valid?
	// maybe we do not actually need to do this.

	jsonFile, err := os.Open(PathToFeeds)
	if err != nil {
		log.Fatal(err)
	}

	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &feeds)

	for index, feedName := range feeds {
		if feedName == opts["feed"].StringValue() {
			// remove feed from slice
			feeds = append(feeds[:index], feeds[index+1:]...)
			byteValue, err = json.Marshal(feeds)
			if err != nil {
				log.Fatal(err)
			}
			err = os.WriteFile(PathToFeeds, byteValue, 0644)
			if err != nil {
				log.Fatal(err)
			}

			builder := new(strings.Builder)
			builder.WriteString("Feed URL: `" + opts["feed"].StringValue() + "` was successfully removed!")

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: builder.String(),
				},
			})
			if err != nil {
				log.Panicf("could not respond to interaction: %s", err)
			}
			return
		}
	}

	builder := new(strings.Builder)
	builder.WriteString("Failed to find URL: `" + opts["feed"].StringValue() + "`. Aborting...")

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
		},
	})
	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}

func handleRssAddFeed(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {

	//does rss feed already exist?
	var feeds []string

	//is the rss feed valid?
	// maybe we do not actually need to do this.

	jsonFile, err := os.Open(PathToFeeds)
	if err != nil {
		log.Fatal(err)
	}

	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &feeds)

	for _, feedName := range feeds {
		if feedName == opts["feed"].StringValue() {
			fmt.Printf("Feed already exists!")

			builder := new(strings.Builder)
			builder.WriteString("RSS feed URL: `" + opts["feed"].StringValue() + "` is already registered! Aborting...")

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: builder.String(),
				},
			})
			if err != nil {
				log.Panicf("could not respond to interaction: %s", err)
			}
			return
		}
	}
	feeds = append(feeds, opts["feed"].StringValue())
	byteValue, err = json.Marshal(feeds)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(PathToFeeds, byteValue, 0644)
	if err != nil {
		log.Fatal(err)
	}

	builder := new(strings.Builder)
	builder.WriteString("RSS feed URL: `" + opts["feed"].StringValue() + "` was successfully registered!")

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
		},
	})
	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
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
	case "rss_add_feed":
		handleRssAddFeed(s, i, parseOptions(data.Options))
	case "rss_remove_feed":
		handleRssRemoveFeed(s, i, parseOptions(data.Options))
	case "rss_list_show":
		handleRssListShow(s, i)
	default:
		return
	}
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Make sure that the bot is NOT list_showto its own messages
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
}

func init() { flag.Parse() }
func Start() {

	// Create a new Discord session with bot token.
	var err error
	session, err := discordgo.New("Bot " + *Token)
	if err != nil {
		log.Fatalf("Error creating Discord session, %v", err)
	}

	// Set BotID variableuuu
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
