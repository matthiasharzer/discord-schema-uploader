package discord

import (
	"github.com/matthiasharzer/discord-schema-uploader/discord/command"
)

func buildCommands(schemaDirectory string) map[string]command.Command {
	return map[string]command.Command{
		"list":     command.ListCommand(schemaDirectory),
		"download": command.DownloadCommand(schemaDirectory),
	}
}
