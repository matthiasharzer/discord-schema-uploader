package discord

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/matthiasharzer/discord-schema-uploader/discord/command"
	"github.com/matthiasharzer/service-kit/logging"
)

const maxFileSize = 10 * 1024 * 1024 // 10 MB limit
const commandPrefix = "!"

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

func parseCommand(message string) (string, []string) {
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return "", nil
	}
	if strings.HasPrefix(parts[0], commandPrefix) {
		parts[0] = strings.TrimPrefix(parts[0], commandPrefix)
	}
	return parts[0], parts[1:]
}

type schemaChannelHandler struct {
	dg        *discordgo.Session
	channelID string
	schemaDir string
	commands  map[string]command.Command
}

func newSchemaChannelHandler(dg *discordgo.Session, channelID, schemaDir string) *schemaChannelHandler {
	return &schemaChannelHandler{
		dg:        dg,
		channelID: channelID,
		schemaDir: schemaDir,
		commands:  buildCommands(schemaDir),
	}
}

func (h *schemaChannelHandler) handleAttachmentMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
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
		destPath := filepath.Join(h.schemaDir, safeName)

		if err := downloadFile(att.URL, destPath); err != nil {
			logging.Error("error downloading file", "url", att.URL, "file_name", safeName, "err", err)
			_, err = s.ChannelMessageSendReply(m.ChannelID, "❌ Failed to save schematic.", m.Reference())
			if err != nil {
				logging.Warn("error sending download error message", "channel", m.ChannelID, "message", m.ID, "error", err)
			}
			continue
		}

		msg := fmt.Sprintf("✅ Saved **`%s`**! Load it in-game with:\n```//schem load %s```", safeName, safeName)
		_, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		if err != nil {
			logging.Warn("error sending save command message", "channel", m.ChannelID, "message", m.ID, "error", err)
		}
	}
}

func (h *schemaChannelHandler) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.ChannelID != h.channelID {
		return
	}

	logging.Info("discord message received", "content", m.Content, "channel", m.ChannelID, "num_attachments", len(m.Attachments))

	if len(m.Attachments) > 0 {
		h.handleAttachmentMessage(s, m)
		return
	}

	if !strings.HasPrefix(m.Content, commandPrefix) {
		return
	}

	cmd, args := parseCommand(m.Content)
	commandFunc, exists := h.commands[cmd]
	if !exists {
		_, err := s.ChannelMessageSendReply(m.ChannelID, "❌ Unknown command.", m.Reference())
		if err != nil {
			logging.Warn("error sending unknown command message", "channel", m.ChannelID, "message", m.ID, "error", err)
		}
		return
	}

	err := commandFunc(args, s, m.Message)
	if err != nil {
		logging.Error("error executing command", "command", cmd, "error", err)
		_, sendErr := s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("❌ Error executing command: %v", err), m.Reference())
		if sendErr != nil {
			logging.Warn("error sending command execution error message", "channel", m.ChannelID, "message", m.ID, "error", sendErr)
		}
	}
}

func (h *schemaChannelHandler) registerHandler() {
	h.dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent
	h.dg.AddHandler(h.handleMessage)
}
