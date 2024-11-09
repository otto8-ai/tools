package id

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/gptscript-ai/go-gptscript"
)

type Cache struct {
	OutlookToNumber map[string]int `json:"outlookToNumber"`
	NumberToOutlook map[int]string `json:"numberToOutlook"`
}

const cacheLocation = "outlookcache.json"

func loadCache(ctx context.Context, gs *gptscript.GPTScript) (Cache, error) {
	cacheBytes, err := gs.ReadFileInWorkspace(ctx, cacheLocation)
	if err != nil {
		var notFoundError *gptscript.NotFoundInWorkspaceError
		if errors.As(err, &notFoundError) {
			return Cache{
				OutlookToNumber: make(map[string]int),
				NumberToOutlook: make(map[int]string),
			}, nil
		}
		return Cache{}, fmt.Errorf("failed to read the Outlook cache file: %w", err)
	}

	var cache Cache
	if err = json.Unmarshal(cacheBytes, &cache); err != nil {
		return Cache{}, fmt.Errorf("failed to unmarshal the Outlook cache: %w", err)
	}
	return cache, nil
}

func writeCache(ctx context.Context, gs *gptscript.GPTScript, c Cache) error {
	cacheBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal the Outlook cache: %w", err)
	}

	if err = gs.WriteFileInWorkspace(ctx, cacheLocation, cacheBytes); err != nil {
		return fmt.Errorf("failed to write the Outlook cache file: %w", err)
	}
	return nil
}

func GetOutlookID(ctx context.Context, id string) (string, error) {
	ids, err := GetOutlookIDs(ctx, []string{id})
	if err != nil {
		return "", err
	}
	return ids[id], nil
}

func GetOutlookIDs(ctx context.Context, ids []string) (map[string]string, error) {
	if len(ids) == 0 {
		return map[string]string{}, nil
	}

	gs, err := gptscript.NewGPTScript()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the GPTScript client: %w", err)
	}

	cache, err := loadCache(ctx, gs)
	if err != nil {
		return nil, err
	}

	results := map[string]string{}
	for _, id := range ids {
		idNum, err := strconv.Atoi(id)
		if err != nil {
			// If the ID does not convert to a number, it's most likely already an Outlook ID, so we just return it back.
			results[id] = id
			continue
		}

		outlookID, ok := cache.NumberToOutlook[idNum]
		if !ok {
			return nil, fmt.Errorf("error: Outlook ID not found")
		}

		results[id] = outlookID
	}

	return results, nil
}

func SetOutlookID(ctx context.Context, outlookID string) (string, error) {
	ids, err := SetOutlookIDs(ctx, []string{outlookID})
	if err != nil {
		return "", err
	}
	return ids[outlookID], nil
}

func SetOutlookIDs(ctx context.Context, outlookIDs []string) (map[string]string, error) {
	if len(outlookIDs) == 0 {
		return map[string]string{}, nil
	}

	gs, err := gptscript.NewGPTScript()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the GPTScript client: %w", err)
	}

	cache, err := loadCache(ctx, gs)
	if err != nil {
		return nil, err
	}

	results := map[string]string{}
	mustWrite := false
	for _, outlookID := range outlookIDs {
		// First we try looking for an existing one.
		numID, ok := cache.OutlookToNumber[outlookID]

		// If it doesn't exist, we create a new one.
		if !ok {
			numID = len(cache.OutlookToNumber) + 1
			cache.OutlookToNumber[outlookID] = numID
			cache.NumberToOutlook[numID] = outlookID
			mustWrite = true
		}

		results[outlookID] = strconv.Itoa(numID)
	}

	if mustWrite {
		if err = writeCache(ctx, gs, cache); err != nil {
			return nil, err
		}
	}

	return results, nil
}
