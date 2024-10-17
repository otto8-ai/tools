package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	url2 "net/url"
	"os"
	"path"
	"strings"
	"time"

	"fmt"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
)

type Colly struct {
	collector *colly.Collector
}

func NewColly() *Colly {
	return &Colly{
		collector: colly.NewCollector(),
	}
}

func (c *Colly) Crawl(metadata *Metadata, metadataPath string, workingDir string) {
	converter := md.NewConverter("", true, nil)

	visited := make(map[string]struct{})
	folders := make(map[string]struct{})
	exclude := make(map[string]bool)

	for _, url := range metadata.Input.Exclude {
		exclude[url] = true
	}

	for _, url := range metadata.Input.WebsiteCrawlingConfig.URLs {
		c.collector.OnHTML("body", func(e *colly.HTMLElement) {
			if _, ok := visited[e.Request.URL.String()]; ok {
				return
			}
			if exclude[e.Request.URL.String()] {
				return
			}

			visited[e.Request.URL.String()] = struct{}{}
			markdown := converter.Convert(e.DOM)
			hostname := e.Request.URL.Hostname()
			urlPath := e.Request.URL.Path

			var filePath string
			if urlPath == "" {
				filePath = path.Join(workingDir, hostname, "index.md")
			} else {
				trimmedPath := strings.Trim(urlPath, "/")
				if trimmedPath == "" {
					filePath = path.Join(workingDir, hostname, "index.md")
				} else {
					segments := strings.Split(trimmedPath, "/")
					fileName := segments[len(segments)-1] + ".md"
					filePath = path.Join(path.Join(workingDir, hostname, strings.Join(segments[:len(segments)-1], "/")), fileName)
				}
			}
			dirPath := path.Dir(filePath)
			err := os.MkdirAll(dirPath, os.ModePerm)
			if err != nil {
				logrus.Errorf("Failed to create directories for %s: %v", dirPath, err)
				return
			}

			etag := e.Response.Headers.Get("ETag")
			lastModified := e.Response.Headers.Get("Last-Modified")
			var updatedAt string
			if etag != "" {
				updatedAt = etag
			} else if lastModified != "" {
				updatedAt = lastModified
			} else {
				updatedAt = time.Now().Format(time.RFC3339)
			}

			checksum, err := GetChecksum(markdown)
			if err != nil {
				logrus.Errorf("Failed to get checksum for %s: %v", filePath, err)
				return
			}
			if checksum == metadata.Output.Files[e.Request.URL.String()].Checksum {
				logrus.Infof("skipping %s because it has not changed", e.Request.URL.String())
				return
			}

			if updatedAt == metadata.Output.Files[e.Request.URL.String()].UpdatedAt {
				logrus.Infof("skipping %s because it has not changed for etag/last-modified: %s/%s", e.Request.URL.String(), etag, lastModified)
				return
			}

			logrus.Infof("scraping %s", e.Request.URL.String())
			err = os.WriteFile(filePath, []byte(markdown), 0644)
			if err != nil {
				logrus.Errorf("Failed to write markdown to %s: %v", filePath, err)
				return
			}
			visited[e.Request.URL.String()] = struct{}{}

			metadata.Output.Files[e.Request.URL.String()] = FileDetails{
				FilePath:  filePath,
				URL:       e.Request.URL.String(),
				UpdatedAt: updatedAt,
				Checksum:  checksum,
			}

			folders[path.Join(workingDir, hostname)] = struct{}{}
			metadata.Output.State.WebsiteCrawlingState.Folders = folders

			metadata.Output.Status = fmt.Sprintf("scraped %d pages", len(visited))
			if err := writeMetadata(metadata, metadataPath); err != nil {
				logrus.Fatalf("Failed to write metadata: %v", err)
			}
		})

		c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
			link := e.Attr("href")

			baseURL, err := url2.Parse(url)
			if err != nil {
				logrus.Errorf("Invalid base URL: %v", err)
				return
			}
			linkURL, err := url2.Parse(link)
			if err != nil {
				logrus.Errorf("Invalid link URL %s: %v", link, err)
				return
			}
			if _, ok := visited[linkURL.String()]; ok {
				return
			}
			if strings.ToLower(path.Ext(linkURL.Path)) == ".pdf" {
				if err := scrapePDF(workingDir, metadata, metadataPath, visited, exclude, linkURL, baseURL); err != nil {
					logrus.Errorf("Failed to scrape PDF %s: %v", linkURL.String(), err)
				}
			} else if (linkURL.Host == "" || baseURL.Host == linkURL.Host) && strings.HasPrefix(linkURL.Path, baseURL.Path) {
				if linkURL.Host == "" && !strings.HasPrefix(link, "#") {
					fullLink := baseURL.ResolveReference(linkURL).String()
					parsedLink, err := url2.Parse(fullLink)
					if err != nil {
						logrus.Errorf("Invalid link URL %s: %v", link, err)
						return
					}
					// don't scrape duplicate pages for homepage, for example, https://www.acorn.io and https://www.acorn.io/
					if parsedLink.Path == "/" {
						parsedLink.Path = ""
					}
					e.Request.Visit(parsedLink.String())
				} else if !strings.HasPrefix(link, "#") {
					e.Request.Visit(linkURL.String())
				}
			}
		})

		err := c.collector.Visit(url)
		if err != nil {
			logrus.Errorf("Failed to visit %s: %v", url, err)
			metadata.Output.Error = err.Error()
		}
	}
	for url, file := range metadata.Output.Files {
		if _, ok := visited[url]; !ok || exclude[url] {
			logrus.Infof("removing file %s", file.FilePath)
			if err := os.RemoveAll(file.FilePath); err != nil {
				logrus.Errorf("Failed to remove %s: %v", file.FilePath, err)
			}
			delete(metadata.Output.Files, url)
		}
	}

	for folder := range metadata.Output.State.WebsiteCrawlingState.Folders {
		if _, ok := folders[folder]; !ok {
			logrus.Infof("removing folder %s", folder)
			if err := os.RemoveAll(folder); err != nil {
				logrus.Errorf("Failed to remove %s: %v", folder, err)
			}
			delete(metadata.Output.State.WebsiteCrawlingState.Folders, folder)
		}
	}

	metadata.Output.State.WebsiteCrawlingState.Pages = make(map[string]struct{})
	for url := range metadata.Output.Files {
		metadata.Output.State.WebsiteCrawlingState.Pages[url] = struct{}{}
	}

	metadata.Output.Status = ""
	metadata.Output.Error = ""
	if err := writeMetadata(metadata, metadataPath); err != nil {
		logrus.Fatalf("Failed to write metadata: %v", err)
	}
}

func scrapePDF(workingDir string, metadata *Metadata, metadataPath string, visited map[string]struct{}, exclude map[string]bool, linkURL *url2.URL, baseURL *url2.URL) error {
	if linkURL.Host == "" {
		var err error
		fullLink := baseURL.ResolveReference(linkURL).String()
		linkURL, err = url2.Parse(fullLink)
		if err != nil {
			return fmt.Errorf("invalid link URL %s: %v", fullLink, err)
		}
	}
	if _, ok := visited[linkURL.String()]; ok {
		return nil
	}
	if exclude[linkURL.String()] {
		return nil
	}
	logrus.Infof("downloading PDF %s", linkURL.String())
	filePath := path.Join(workingDir, baseURL.Host, linkURL.Host, strings.TrimPrefix(linkURL.Path, "/"))
	dirPath := path.Dir(filePath)
	resp, err := http.Get(linkURL.String())
	if err != nil {
		return fmt.Errorf("failed to download PDF %s: %v", linkURL.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download PDF %s: status code %d", linkURL.String(), resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "temp_pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy PDF to temp file: %v", err)
	}
	tempFile.Seek(0, 0)
	newChecksum, err := GetChecksum(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %v", err)
	}

	if fileDetails, exists := metadata.Output.Files[linkURL.String()]; exists {
		if fileDetails.Checksum == newChecksum {
			logrus.Infof("PDF %s has not been modified", linkURL.String())
			return nil
		}
	}

	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directories for %s: %v", dirPath, err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filePath, err)
	}
	defer file.Close()
	tempFile.Seek(0, 0)
	_, err = io.Copy(file, tempFile)
	if err != nil {
		return fmt.Errorf("failed to save PDF to %s: %v", filePath, err)
	}
	visited[linkURL.String()] = struct{}{}

	metadata.Output.Status = fmt.Sprintf("scraped %d pages", len(visited))
	metadata.Output.Files[linkURL.String()] = FileDetails{
		FilePath:  filePath,
		URL:       linkURL.String(),
		UpdatedAt: time.Now().String(),
		Checksum:  newChecksum,
	}

	if err := writeMetadata(metadata, metadataPath); err != nil {
		return fmt.Errorf("failed to write metadata: %v", err)
	}
	return nil
}

func GetChecksum(content string) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(content))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
