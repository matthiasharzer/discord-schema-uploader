package command

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func addFileToZip(zipWriter *zip.Writer, filePath, fileName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	w, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, file)
	return err
}

func buildZipFromSchemas(schemaDirectory string, schemaNames []string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for _, schemaName := range schemaNames {
		err := addFileToZip(zipWriter, filepath.Join(schemaDirectory, schemaName), schemaName)
		if err != nil {
			_ = zipWriter.Close()
			return nil, err
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func getAllSchemaFiles(schemaDirectory string) ([]string, error) {
	entries, err := os.ReadDir(schemaDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var schemaFiles []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if strings.HasPrefix(ext, ".schem") {
			schemaFiles = append(schemaFiles, entry.Name())
		}
	}
	return schemaFiles, nil
}

func DownloadCommand(schemaDirectory string) Command {
	return func(args []string, s *discordgo.Session, m *discordgo.Message) error {
		if len(args) == 0 {
			_, err := s.ChannelMessageSendReply(m.ChannelID, "❌ No schema name provided. Usage: `!download <schema_name> [<schema_name_2>...]`", m.Reference())
			return err
		}

		if len(args) > 1 {
			body, err := buildZipFromSchemas(schemaDirectory, args)
			if err != nil {
				return err
			}
			_, err = s.ChannelFileSend(m.ChannelID, "schemas.zip", body)
			return err
		}
		if args[0] == "all" {
			schemaFiles, err := getAllSchemaFiles(schemaDirectory)
			if err != nil {
				return err
			}
			if len(schemaFiles) == 0 {
				_, err := s.ChannelMessageSendReply(m.ChannelID, "❌ No schema files found to download.", m.Reference())
				return err
			}
			body, err := buildZipFromSchemas(schemaDirectory, schemaFiles)
			if err != nil {
				return err
			}
			_, err = s.ChannelFileSend(m.ChannelID, "all_schemas.zip", body)
			return err
		}

		schemaPath := filepath.Join(schemaDirectory, args[0])
		file, err := os.Open(schemaPath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = s.ChannelFileSend(m.ChannelID, args[0], file)
		if err != nil {
			return err
		}
		return nil
	}
}
