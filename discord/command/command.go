package command

import "github.com/bwmarrin/discordgo"

type Command = func(args []string, s *discordgo.Session, m *discordgo.Message) error
