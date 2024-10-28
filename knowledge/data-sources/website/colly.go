package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"os"
	"path"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/sirupsen/logrus"
)

func CrawlColly(ctx context.Context, input *MetadataInput, output *MetadataOutput, logErr *logrus.Logger, gptscript *gptscript.GPTScript) error {
	converter := md.NewConverter("", true, nil)

	visited := make(map[string]struct{})
	folders := make(map[string]struct{})

	for _, url := range input.WebsiteCrawlingConfig.URLs {
		if err := scrape(ctx, converter, logErr, output, gptscript, visited, folders, url); err != nil {
			logrus.Errorf("Failed to scrape %s: %v", url, err)
		}
	}

	for url, file := range output.Files {
		if _, ok := visited[url]; !ok {
			logErr.Infof("removing file %s", file.FilePath)
			if err := gptscript.DeleteFileInWorkspace(ctx, file.FilePath); err != nil {
				return err
			}
			delete(output.Files, url)
			delete(output.State.WebsiteCrawlingState.Pages, url)
		}
	}

	for folder := range output.State.WebsiteCrawlingState.Folders {
		if _, ok := folders[folder]; !ok {
			logrus.Infof("removing folder %s", folder)
			if err := os.RemoveAll(folder); err != nil {
				logrus.Errorf("Failed to remove %s: %v", folder, err)
			}
			delete(output.State.WebsiteCrawlingState.Folders, folder)
		}
	}

	output.Status = ""
	return writeMetadata(ctx, output, gptscript)
}

func scrape(ctx context.Context, converter *md.Converter, logErr *logrus.Logger, output *MetadataOutput, gptscript *gptscript.GPTScript, visited map[string]struct{}, folders map[string]struct{}, url string) error {
	collector := colly.NewCollector()
	collector.OnHTML("body", func(e *colly.HTMLElement) {
		if _, ok := visited[e.Request.URL.String()]; ok {
			return
		}

		logErr.Infof("scraping %s", e.Request.URL.String())
		visited[e.Request.URL.String()] = struct{}{}
		markdown := converter.Convert(e.DOM)
		hostname := e.Request.URL.Hostname()
		urlPath := e.Request.URL.Path

		var filePath string
		if urlPath == "" {
			filePath = path.Join(hostname, "index.md")
		} else {
			trimmedPath := strings.Trim(urlPath, "/")
			if trimmedPath == "" {
				filePath = path.Join(hostname, "index.md")
			} else {
				segments := strings.Split(trimmedPath, "/")
				fileName := segments[len(segments)-1] + ".md"
				filePath = path.Join(hostname, strings.Join(segments[:len(segments)-1], "/"), fileName)
			}
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

		checksum, err := getChecksum(markdown)
		if err != nil {
			logErr.Errorf("Failed to get checksum for %s: %v", e.Request.URL.String(), err)
			return
		}
		if checksum == output.Files[e.Request.URL.String()].Checksum {
			logErr.Infof("skipping %s because it has not changed", e.Request.URL.String())
			return
		}

		if updatedAt == output.Files[e.Request.URL.String()].UpdatedAt {
			logErr.Infof("skipping %s because it has not changed for etag/last-modified: %s/%s", e.Request.URL.String(), etag, lastModified)
			return
		}

		if err := gptscript.WriteFileInWorkspace(ctx, filePath, []byte(markdown)); err != nil {
			logErr.Errorf("Failed to write file %s: %v", filePath, err)
			return
		}

		visited[e.Request.URL.String()] = struct{}{}

		output.Files[e.Request.URL.String()] = FileDetails{
			FilePath:  filePath,
			URL:       e.Request.URL.String(),
			UpdatedAt: updatedAt,
			Checksum:  checksum,
		}

		output.State.WebsiteCrawlingState.Pages[e.Request.URL.String()] = PageDetails{
			ParentURL: url,
		}

		folders[hostname] = struct{}{}
		output.State.WebsiteCrawlingState.Folders = folders

		output.Status = fmt.Sprintf("scraped %d pages", len(visited))
		if err := writeMetadata(ctx, output, gptscript); err != nil {
			logrus.Fatalf("Failed to write metadata: %v", err)
		}
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		baseURL, err := url2.Parse(url)
		if err != nil {
			logErr.Errorf("Invalid base URL: %v", err)
			return
		}
		linkURL, err := url2.Parse(link)
		if err != nil {
			logErr.Errorf("Invalid link URL %s: %v", link, err)
			return
		}
		if _, ok := visited[linkURL.String()]; ok {
			return
		}
		if strings.ToLower(path.Ext(linkURL.Path)) == ".pdf" {
			if err := scrapePDF(ctx, output, visited, linkURL, baseURL, gptscript); err != nil {
				logErr.Errorf("Failed to scrape PDF %s: %v", linkURL.String(), err)
			}
		} else if (linkURL.Host == "" || baseURL.Host == linkURL.Host) && strings.HasPrefix(linkURL.Path, baseURL.Path) {
			if linkURL.Host == "" && !strings.HasPrefix(link, "#") {
				fullLink := baseURL.ResolveReference(linkURL).String()
				parsedLink, err := url2.Parse(fullLink)
				if err != nil {
					logErr.Errorf("Invalid link URL %s: %v", link, err)
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
	return collector.Visit(url)
}

func scrapePDF(ctx context.Context, output *MetadataOutput, visited map[string]struct{}, linkURL *url2.URL, baseURL *url2.URL, gptscript *gptscript.GPTScript) error {
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
	logrus.Infof("downloading PDF %s", linkURL.String())
	filePath := path.Join(baseURL.Host, linkURL.Host, strings.TrimPrefix(linkURL.Path, "/"))
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
	newChecksum, err := getChecksum(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %v", err)
	}

	if fileDetails, exists := output.Files[linkURL.String()]; exists {
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

	output.Status = fmt.Sprintf("scraped %d pages", len(visited))
	output.Files[linkURL.String()] = FileDetails{
		FilePath:  filePath,
		URL:       linkURL.String(),
		UpdatedAt: time.Now().String(),
		Checksum:  newChecksum,
	}

	if err := writeMetadata(ctx, output, gptscript); err != nil {
		return fmt.Errorf("failed to write metadata: %v", err)
	}
	return nil
}

func getChecksum(content string) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(content))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
