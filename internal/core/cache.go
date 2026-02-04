package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// TaskCache represents a cached task location
type TaskCache struct {
	GlobalID    int    `json:"global_id"`
	FilePath    string `json:"file_path"`
	LocalID     int    `json:"local_id"`
	TaskText    string `json:"task_text"`
	LastListing time.Time `json:"last_listing"`
}

// Cache represents the task cache
type Cache struct {
	Tasks       map[int]TaskCache `json:"tasks"`
	LastListing time.Time         `json:"last_listing"`
}

// GetCacheFilePath returns the path to the cache file
func GetCacheFilePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	// Use .mdd-cache in the current directory
	return filepath.Join(cwd, ".mdd-cache"), nil
}

// SaveCache saves the cache to disk
func SaveCache(cache *Cache) error {
	cachePath, err := GetCacheFilePath()
	if err != nil {
		return err
	}
	
	cache.LastListing = time.Now()
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	
	return os.WriteFile(cachePath, data, 0644)
}

// LoadCache loads the cache from disk
func LoadCache() (*Cache, error) {
	cachePath, err := GetCacheFilePath()
	if err != nil {
		return nil, err
	}
	
	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, nil // No cache exists
	}
	
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	
	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	
	return &cache, nil
}

// ClearCache removes the cache file
func ClearCache() error {
	cachePath, err := GetCacheFilePath()
	if err != nil {
		return err
	}
	
	// Remove the cache file if it exists
	if _, err := os.Stat(cachePath); err == nil {
		return os.Remove(cachePath)
	}
	
	return nil
}
