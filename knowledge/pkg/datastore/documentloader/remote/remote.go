package documentloader

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func IsRemote(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return true
	}
	return false
}

func LoadRemote(path string) (string, error) {
	slog.Debug("Loading remote path", "path", path)
	tmpDir, err := os.MkdirTemp(os.TempDir(), "remote")
	if err != nil {
		return "", err
	}

	if strings.Contains(path, "github.com") || strings.Contains(path, "gitlab.com") {
		return tmpDir, CloneRepo(path, tmpDir)
	}

	return "", fmt.Errorf("unsupported remote repository %q", path)
}
