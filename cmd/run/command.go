package run

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/matthiasharzer/discord-schema-uploader/discord"
	"github.com/matthiasharzer/service-kit/logging"
	"github.com/spf13/cobra"
)

var channelID string
var schemaDir string

func init() {
	Command.Flags().StringVarP(&channelID, "channel-id", "c", "", "Discord channel ID to listen for schematics")
	Command.Flags().StringVarP(&schemaDir, "schema-dir", "s", "", "Directory to put schematic in")

	err := Command.MarkFlagRequired("channel-id")
	if err != nil {
		panic(err)
	}
	err = Command.MarkFlagRequired("schema-dir")
	if err != nil {
		panic(err)
	}
}

var Command = &cobra.Command{
	Use: "run",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		info, err := os.Stat(schemaDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("schema directory %s does not exist", schemaDir)
			}
			return err
		}
		if !info.IsDir() {
			return errors.New("schema directory is not a directory")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		discordToken := os.Getenv("DISCORD_TOKEN")
		if discordToken == "" {
			return errors.New("DISCORD_TOKEN environment variable must be set")
		}

		bot, err := discord.NewSchemaBot(discordToken, channelID, schemaDir)
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
