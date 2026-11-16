package discord

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/matthiasharzer/service-kit/logging"
)

const maxFileSize = 10 * 1024 * 1024 // 10 MB limit

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, io.LimitReader(resp.Body, maxFileSize))
	return err
}

type SchemaBot struct {
	dg        *discordgo.Session
	channelID string
	schemaDir string
}

func NewSchemaBot(discordToken, channelID, schemaDir string) (*SchemaBot, error) {
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %v", err)
	}
	return &SchemaBot{
		dg:        dg,
		channelID: channelID,
		schemaDir: schemaDir,
	}, nil
}

func (b *SchemaBot) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.ChannelID != b.channelID || len(m.Attachments) == 0 {
		return
	}

	logging.Info("discord message received", "content", m.Content, "channel", m.ChannelID, "num_attachments", len(m.Attachments))

	for _, att := range m.Attachments {
		ext := strings.ToLower(filepath.Ext(att.Filename))
		if ext != ".schem" && ext != ".schematic" {
			_, err := s.ChannelMessageSendReply(m.ChannelID, "❌ Not a schematic file.", m.Reference())
			if err != nil {
				logging.Warn("error sending file type error message", "channel", m.ChannelID, "message", m.ID, "error", err)
			}
			continue
		}

		if att.Size > maxFileSize {
			_, err := s.ChannelMessageSendReply(m.ChannelID, "❌ File exceeds 10 MB limit.", m.Reference())
			if err != nil {
				logging.Warn("error sending file size exceeded message", "channel", m.ChannelID, "message", m.ID, "error", err)
			}
			continue
		}

		safeName := filepath.Base(att.Filename)
		destPath := filepath.Join(b.schemaDir, safeName)

		if err := downloadFile(att.URL, destPath); err != nil {
			logging.Error("error downloading file", "url", att.URL, "file_name", safeName, "err", err)
			_, err = s.ChannelMessageSendReply(m.ChannelID, "❌ Failed to save schematic.", m.Reference())
			if err != nil {
				logging.Warn("error sending download error message", "channel", m.ChannelID, "message", m.ID, "error", err)
			}
			continue
		}

		cmdName := strings.TrimSuffix(safeName, ext)
		msg := fmt.Sprintf("✅ Saved **`%s`**! Load it in-game with:\n```//schem load %s```", safeName, cmdName)
		_, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		if err != nil {
			logging.Warn("error sending save command message", "channel", m.ChannelID, "message", m.ID, "error", err)
		}
	}
}

func (b *SchemaBot) Start() error {
	b.dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent
	b.dg.AddHandler(b.handleMessage)
	return b.dg.Open()
}

func (b *SchemaBot) Close() error {
	return b.dg.Close()
}
