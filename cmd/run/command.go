package run

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/matthiasharzer/discord-schema-uploader/discord"
	"github.com/matthiasharzer/service-kit/logging"
	"github.com/spf13/cobra"
)

func parseHandlerArg(arg string) (discord.SchemaHandlerConfig, error) {
	parts := strings.Split(arg, ":")
	if len(parts) != 2 {
		return discord.SchemaHandlerConfig{}, errors.New("invalid handler argument format. Expected: channel_id:schema_dir")
	}
	channelID := parts[0]
	schemaDirectory := parts[1]
	if channelID == "" || schemaDirectory == "" {
		return discord.SchemaHandlerConfig{}, errors.New("channel_id and schema_dir cannot be empty")
	}
	info, err := os.Stat(schemaDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			return discord.SchemaHandlerConfig{}, fmt.Errorf("schema directory %s does not exist", schemaDirectory)
		}
		return discord.SchemaHandlerConfig{}, err
	}
	if !info.IsDir() {
		return discord.SchemaHandlerConfig{}, errors.New("schema directory is not a directory")
	}
	return discord.SchemaHandlerConfig{
		ChannelID: channelID,
		SchemaDir: schemaDirectory,
	}, nil
}

var Command = &cobra.Command{
	Use:  "run",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var handlers []discord.SchemaHandlerConfig
		for _, arg := range args {
			h, err := parseHandlerArg(arg)
			if err != nil {
				return fmt.Errorf("error parsing handler argument %s: %v", arg, err)
			}
			handlers = append(handlers, h)
		}

		discordToken := os.Getenv("DISCORD_TOKEN")
		if discordToken == "" {
			return errors.New("DISCORD_TOKEN environment variable must be set")
		}

		bot, err := discord.NewSchemaBot(discordToken, handlers)
		if err != nil {
			return err
		}
		err = bot.Start()
		if err != nil {
			return err
		}

		defer func() {
			err := bot.Close()
			if err != nil {
				logging.Error("error stopping discord bot", "err", err)
			}
		}()

		logging.Info("discord bot is now running. Press CTRL-C to exit.")
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-stop

		return nil
	},
}
