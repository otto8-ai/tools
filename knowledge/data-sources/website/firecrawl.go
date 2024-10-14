package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mendableai/firecrawl-go"
	"github.com/sirupsen/logrus"
)

type Firecrawl struct {
	app *firecrawl.FirecrawlApp
}

func NewFirecrawl() *Firecrawl {
	firecrawlUrl := os.Getenv("FIRECRAWL_URL")
	if firecrawlUrl == "" {
		firecrawlUrl = "http://localhost:3002"
	}
	apiKey := os.Getenv("FIRECRAWL_API_KEY")
	if apiKey == "" {
		apiKey = "test"
	}
	app, err := firecrawl.NewFirecrawlApp(apiKey, firecrawlUrl)
	if err != nil {
		log.Fatalf("Failed to initialize FirecrawlApp: %v", err)
	}
	return &Firecrawl{app: app}
}

func (f *Firecrawl) Crawl(metadata *Metadata, metadataPath string, workingDir string) {
	if metadata.Output.State.WebsiteCrawlingState.ScrapeJobIds == nil {
		metadata.Output.State.WebsiteCrawlingState.ScrapeJobIds = make(map[string]string)
	}
	if metadata.Output.State.WebsiteCrawlingState.Folders == nil {
		metadata.Output.State.WebsiteCrawlingState.Folders = make(map[string]struct{})
	}

	for url := range metadata.Output.State.WebsiteCrawlingState.ScrapeJobIds {
		found := false
		for _, inputURL := range metadata.Input.WebsiteCrawlingConfig.URLs {
			if inputURL == url {
				found = true
				break
			}
		}
		if !found {
			delete(metadata.Output.State.WebsiteCrawlingState.ScrapeJobIds, url)
		}
	}

	folders := make(map[string]struct{})
	visitedPages := make(map[string]struct{})

	for _, url := range metadata.Input.WebsiteCrawlingConfig.URLs {
		if metadata.Output.State.WebsiteCrawlingState.ScrapeJobIds[url] == "" {
			crawlStatus, err := f.app.AsyncCrawlURL(url, &firecrawl.CrawlParams{
				Limit: &[]int{100}[0],
				ScrapeOptions: firecrawl.ScrapeParams{
					Formats: []string{"markdown"},
				},
			}, nil)
			if err != nil {
				log.Fatalf("Failed to send crawl request: %v", err)
			}
			metadata.Output.State.WebsiteCrawlingState.ScrapeJobIds[url] = crawlStatus.ID
			if err := writeMetadata(metadata, metadataPath); err != nil {
				log.Fatalf("Failed to write metadata: %v", err)
			}
		}
	}

	for _, scrapeJobId := range metadata.Output.State.WebsiteCrawlingState.ScrapeJobIds {
		fileWritten := 0
	OuterLoop:
		for {
			crawlStatus, err := f.app.CheckCrawlStatus(scrapeJobId)
			if err != nil {
				log.Fatalf("Failed to check crawl status: %v", err)
			}
			if crawlStatus.Status == "completed" {
				for {
					for _, data := range crawlStatus.Data {
						source := ""
						if data.Metadata.SourceURL != nil {
							source = *data.Metadata.SourceURL
						}
						lastModified := time.Now().String()
						if data.Metadata.ModifiedTime != nil {
							lastModified = *data.Metadata.ModifiedTime
						}

						parsedURL, err := url.Parse(source)
						if err != nil {
							logrus.Errorf("Failed to parse URL %s: %v", source, err)
							continue
						}

						urlPath := ""
						if parsedURL.Path == "" || parsedURL.Path == "/" {
							urlPath = "/index"
						} else {
							urlPath = parsedURL.Path
						}

						existingPage, ok := metadata.Output.Files[source]
						if ok && existingPage.UpdatedAt == lastModified && lastModified != "" {
							continue
						}

						markdownPath := filepath.Join(workingDir, parsedURL.Host, urlPath) + ".md"

						if err := os.MkdirAll(filepath.Dir(markdownPath), 0755); err != nil {
							logrus.Errorf("Failed to create directory for %s: %v", markdownPath, err)
							continue
						}

						if err := os.WriteFile(markdownPath, []byte(data.Markdown), 0644); err != nil {
							logrus.Errorf("Failed to write markdown file %s: %v", markdownPath, err)
							continue
						}

						metadata.Output.Files[source] = FileDetails{
							UpdatedAt: lastModified,
							FilePath:  markdownPath,
							URL:       source,
						}
						visitedPages[source] = struct{}{}

						folders[filepath.Join(workingDir, parsedURL.Host)] = struct{}{}
						metadata.Output.State.WebsiteCrawlingState.Folders[filepath.Join(workingDir, parsedURL.Host)] = struct{}{}

						fileWritten++
						logrus.Infof("wrote %d webpages to disk", fileWritten)
						metadata.Output.Status = fmt.Sprintf("wrote %d webpages to disk", fileWritten)

						if err := writeMetadata(metadata, metadataPath); err != nil {
							logrus.Fatalf("Failed to write metadata: %v", err)
						}
					}
					if crawlStatus.Next == nil || *crawlStatus.Next == "" {
						break OuterLoop
					} else {
						if !strings.HasPrefix(*crawlStatus.Next, "http://") {
							*crawlStatus.Next = "http://" + strings.TrimPrefix(*crawlStatus.Next, "https://")
						}

						req, err := http.NewRequest(http.MethodGet, *crawlStatus.Next, nil)
						if err != nil {
							log.Fatalf("Failed to create request: %v", err)
						}
						req.Header.Add("Content-Type", "application/json")
						client := &http.Client{}
						var respBody []byte
						for i := 0; i < 3; i++ {
							resp, err := client.Do(req)
							if err != nil {
								time.Sleep(500 * time.Millisecond)
								continue
							}
							defer resp.Body.Close()
							respBody, err = io.ReadAll(resp.Body)
							if err != nil {
								log.Fatalf("Failed to read response body: %v", err)
							}
							break
						}
						if err != nil {
							log.Fatalf("Failed to get next crawl status: %v", err)
						}
						newCrawlStatus := firecrawl.CrawlStatusResponse{}
						err = json.Unmarshal(respBody, &newCrawlStatus)
						if err != nil {
							log.Fatalf("Failed to unmarshal next crawl status: %v", err)
						}
						crawlStatus = &newCrawlStatus
					}
				}
			} else {
				metadata.Output.Status = fmt.Sprintf("crawling status: %s, completed %d, total %d", crawlStatus.Status, crawlStatus.Completed, crawlStatus.Total)
				if err := writeMetadata(metadata, metadataPath); err != nil {
					log.Fatalf("Failed to write metadata: %v", err)
				}
				continue
			}
		}
	}

	for folder := range metadata.Output.State.WebsiteCrawlingState.Folders {
		if _, ok := folders[folder]; !ok {
			if err := os.RemoveAll(folder); err != nil {
				logrus.Errorf("Failed to remove folder %s: %v", folder, err)
			}
			delete(metadata.Output.State.WebsiteCrawlingState.Folders, folder)
		}
	}

	for page, detail := range metadata.Output.Files {
		if _, ok := visitedPages[page]; !ok {
			if err := os.RemoveAll(detail.FilePath); err != nil {
				logrus.Errorf("Failed to remove page %s: %v", page, err)
			}
			delete(metadata.Output.Files, page)
		}
	}

	metadata.Output.Status = "done"
	if err := writeMetadata(metadata, metadataPath); err != nil {
		log.Fatalf("Failed to write metadata: %v", err)
	}

}
