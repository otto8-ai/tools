package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gptscript-ai/knowledge/pkg/datastore"
	"github.com/gptscript-ai/knowledge/pkg/index"
	"github.com/spf13/cobra"
)

type ClientDeleteFile struct {
	Client
	Dataset string `usage:"Target Dataset ID" short:"d" default:"default"`
}

func (s *ClientDeleteFile) Customize(cmd *cobra.Command) {
	cmd.Use = "delete-file <file-id|file-abs-path>"
	cmd.Short = "Delete a file from a dataset"
	cmd.Args = cobra.ExactArgs(1)
}

func (s *ClientDeleteFile) Run(cmd *cobra.Command, args []string) error {
	c, err := s.getClient(cmd.Context())
	if err != nil {
		return err
	}

	fileRef := args[0]

	searchFile := index.File{
		Dataset: s.Dataset,
	}

	if strings.HasPrefix(fileRef, "/") || strings.HasPrefix(fileRef, "ws://") {
		searchFile.AbsolutePath = fileRef
	} else if _, err := uuid.Parse(fileRef); err == nil {
		searchFile.ID = fileRef
	} else {
		finfo, err := os.Stat(fileRef)
		if err != nil {
			return fmt.Errorf("fileref is not a valid filepath or UUID - failed to stat relative path: %w", err)
		}
		if finfo.IsDir() {
			return fmt.Errorf("fileref is a directory, not a file")
		}
		searchFile.AbsolutePath, err = filepath.Abs(fileRef)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
	}

	file, err := c.FindFile(cmd.Context(), searchFile)
	if errors.Is(err, datastore.ErrDBFileNotFound) || file == nil {
		slog.Info("File not found", "file", searchFile)
		return nil
	}
	if err != nil {
		return err
	}

	err = c.DeleteFile(cmd.Context(), s.Dataset, file.ID)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	fmt.Printf("File %s (%s) deleted\n", file.ID, file.AbsolutePath)

	return nil
}
