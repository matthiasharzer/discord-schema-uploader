package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ListCommand(schemaDir string) Command {
	return func(args []string, s *discordgo.Session, m *discordgo.Message) error {
		entries, err := os.ReadDir(schemaDir)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		var message strings.Builder
		message.WriteString("## Available Schematics\n")

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			message.WriteString(fmt.Sprintf("- `%s`\n", entry.Name()))
		}

		_, err = s.ChannelMessageSendReply(m.ChannelID, message.String(), m.Reference())
		if err != nil {
			return fmt.Errorf("failed to send message reply: %w", err)
		}
		return nil
	}
}
