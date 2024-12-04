# Knowledge Website Integration

## Description

A sync tool to scrape websites and store the content as markdown files for knowledge base system.

## Usage

Instructions on how to use the project.

```bash
go run main.go
```

provide a .metadata.json file to specify the website to be scraped and the local directory to store the markdown files.

```json
{
  "input": {
    "urls": ["https://coral.org"]
  }
}
```