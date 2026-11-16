package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type SchemaBot struct {
	dg       *discordgo.Session
	handlers []schemaChannelHandler
}

func NewSchemaBot(discordToken string, configs []SchemaHandlerConfig) (*SchemaBot, error) {
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %v", err)
	}

	var handlers []schemaChannelHandler
	for _, config := range configs {
		handler := newSchemaChannelHandler(dg, config.ChannelID, config.SchemaDir)
		handlers = append(handlers, *handler)
	}

	return &SchemaBot{
		dg:       dg,
		handlers: handlers,
	}, nil
}

func (b *SchemaBot) Start() error {
	for _, handler := range b.handlers {
		handler.registerHandler()
	}
	return b.dg.Open()
}

func (b *SchemaBot) Close() error {
	return b.dg.Close()
}
