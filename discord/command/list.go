package command

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

func ListCommand(schemaDir string) Command {
	return func(args []string, s *discordgo.Session, m *discordgo.Message) error {
		entries, err := os.ReadDir(schemaDir)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		message := "## Available Schematics\n"

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			message += fmt.Sprintf("- `%s`\n", entry.Name())
		}

		_, err = s.ChannelMessageSendReply(m.ChannelID, message, m.Reference())
		if err != nil {
			return fmt.Errorf("failed to send message reply: %w", err)
		}
		return nil
	}
}
