package converter

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SofficeConverter struct{}

func NewSofficeConverter() (*SofficeConverter, error) {
	if _, err := exec.LookPath("soffice"); err != nil {
		return nil, fmt.Errorf("soffice binary not found")
	}
	return &SofficeConverter{}, nil
}

func (c *SofficeConverter) Convert(ctx context.Context, reader io.Reader, format string) (io.Reader, error) {
	// Convert the file using soffice
	format = strings.ToLower(format)
	ext := format
	switch format {
	case "pdf": // nothing to do
	default:
		return nil, fmt.Errorf("soffice converter - unsupported output format %q", format)
	}

	tempfile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("knowledge-convsource-*.%s", ext))
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

	// Convert the file using soffice
	cmd := exec.Command("soffice", "--headless", "--convert-to", format, "--outdir", os.TempDir(), p)
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	// Open the converted file
	cp := strings.TrimSuffix(p, filepath.Ext(p)) + "." + format

	return os.Open(cp)
}
