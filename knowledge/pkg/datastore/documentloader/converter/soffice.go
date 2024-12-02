package converter

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/log"
)

// compile time check
var _ Converter = (*SofficeConverter)(nil)

type SofficeConverter struct{}

func (c *SofficeConverter) Name() string {
	return "soffice (libreoffice)"
}

func NewSofficeConverter() (*SofficeConverter, error) {
	return &SofficeConverter{}, nil
}

func (c *SofficeConverter) Convert(ctx context.Context, reader io.Reader, sourceExt, outputFormat string) (io.Reader, error) {
	// Convert the file using soffice
	outputFormat = strings.ToLower(outputFormat)
	sourceExt = strings.ToLower(sourceExt)
	switch outputFormat {
	case "pdf": // nothing to do
	default:
		return nil, fmt.Errorf("soffice converter - unsupported output format %q", outputFormat)
	}

	if _, err := exec.LookPath("soffice"); err != nil {
		return nil, fmt.Errorf("soffice binary not found")
	}

	tempfile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("knowledge-convsource-*.%s", strings.TrimPrefix(sourceExt, ".")))
	if err != nil {
		return nil, err
	}
	defer tempfile.Close()

	p := tempfile.Name()

	_, err = io.Copy(tempfile, reader)
	if err != nil {
		return nil, err
	}
	_ = tempfile.Close()

	logger := log.FromCtx(ctx)

	profileDir, err := os.MkdirTemp(os.TempDir(), "libreoffice-profile-*")
	if err != nil {
		logger.Error("Failed to create soffice profile directory", "path", profileDir, "error", err)
		return nil, fmt.Errorf("failed to create soffice profile directory: %w", err)
	}
	defer os.RemoveAll(profileDir)

	// Convert the file using soffice
	cmd := exec.Command(
		"soffice",
		"--headless",
		fmt.Sprintf("-env:UserInstallation=file://%s", profileDir),
		"--convert-to", outputFormat,
		"--outdir", os.TempDir(),
		p,
	)

	// capture stdout and stderr in a buffer
	var outb, errb strings.Builder
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	logger.Debug("Running soffice command", "command", cmd.String())

	err = cmd.Run()
	if err != nil {
		logger.Error("Failed to run soffice command", "error", err, "stderr", errb.String(), "stdout", outb.String())
		return nil, err
	}

	logger.Debug("soffice command output", "stdout", outb.String(), "stderr", errb.String())

	// Open the converted file
	cp := strings.TrimSuffix(p, filepath.Ext(p)) + "." + outputFormat

	return os.Open(cp)
}
