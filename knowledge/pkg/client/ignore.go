package client

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const DefaultIgnoreFile = ".knowignore"

var DefaultIgnorePatterns = []gitignore.Pattern{
	gitignore.ParsePattern(DefaultIgnoreFile, nil), // Default ignore patterns
	gitignore.ParsePattern(MetadataFilename, nil),  // Knowledge Metadata file
	gitignore.ParsePattern("~$*", nil),             // MS Office temp files
	gitignore.ParsePattern("$*", nil),              // Likely hidden/tempfiles
}

func isIgnored(ignore gitignore.Matcher, path string) bool {
	return ignore.Match(strings.Split(path, string(filepath.Separator)), false)
}

func readDefaultIgnoreFile(dirPath string) ([]gitignore.Pattern, error) {
	ignoreFilePath := filepath.Join(dirPath, DefaultIgnoreFile)
	_, err := os.Stat(ignoreFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to check if ignore file %q exists: %w", ignoreFilePath, err)
	}

	return readIgnoreFile(ignoreFilePath)
}

func useDefaultIgnoreFileIfExists(path string) ([]gitignore.Pattern, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	finfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to check if path %q exists: %w", path, err)
	}
	if !finfo.IsDir() {
		path = filepath.Dir(path)
	}

	ignorePatterns, err := readDefaultIgnoreFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read default ignore file: %w", err)
	}

	return ignorePatterns, nil
}

func readIgnoreFile(path string) ([]gitignore.Pattern, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to checkout ignore file %q: %w", path, err)
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("ignore file %q is a directory", path)
	}

	var ps []gitignore.Pattern
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open ignore file %q: %w", path, err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := scanner.Text()
		if !strings.HasPrefix(s, "#") && len(strings.TrimSpace(s)) > 0 {
			ps = append(ps, gitignore.ParsePattern(s, nil))
		}
	}

	return ps, nil
}
