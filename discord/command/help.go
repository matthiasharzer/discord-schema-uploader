package command

import "github.com/bwmarrin/discordgo"

func HelpCommand() Command {
	return func(args []string, s *discordgo.Session, m *discordgo.Message) error {
		helpMessage := "Available commands:\n" +
			"- `!list` - Lists all available schemas.\n" +
			"- `!download <schema_name> [<schema_name_2>...]` - Downloads one or more schemas. Use `all` to download all schemas.\n" +
			"- `!help` - Displays this help message.\n"
		_, err := s.ChannelMessageSend(m.ChannelID, helpMessage)
		return err
	}
}
