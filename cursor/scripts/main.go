package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Cache struct {
	LastVersion string `json:"last_version"`
}

type Monitor struct {
	baseURL    string
	cacheFile  string
	updatePath string
}

func NewMonitor() *Monitor {
	return &Monitor{
		baseURL:    "https://cursor.sh",
		cacheFile:  "cursor/data/last_check.json",
		updatePath: "cursor/updates",
	}
}

func (m *Monitor) fetchCurrentVersion() (string, error) {
	resp, err := http.Get(m.baseURL)
	if err != nil {
		return "", fmt.Errorf("error fetching updates: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	// Note: Replace this with actual selector for version information
	version := doc.Find("version-selector").Text()
	return version, nil
}

func (m *Monitor) createUpdateDirectory() (string, error) {
	today := time.Now().Format("2006-01-02")
	path := filepath.Join(m.updatePath, today)

	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("error creating directory: %v", err)
	}

	readmePath := filepath.Join(path, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		template := fmt.Sprintf(`# %s Updates

## Overview

## Key Features

## Content Planning
- [ ] Technical blog post
- [ ] Feature demonstration video
- [ ] X.com announcement thread

## Resources
- Screenshots (TBD)
- Code examples (TBD)
- Demo videos (TBD)
`, today)

		if err := ioutil.WriteFile(readmePath, []byte(template), 0644); err != nil {
			return "", fmt.Errorf("error creating README: %v", err)
		}
	}

	return path, nil
}

func (m *Monitor) loadCache() (*Cache, error) {
	data, err := ioutil.ReadFile(m.cacheFile)
	if os.IsNotExist(err) {
		return &Cache{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading cache: %v", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("error parsing cache: %v", err)
	}

	return &cache, nil
}

func (m *Monitor) saveCache(version string) error {
	cache := Cache{LastVersion: version}
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("error marshaling cache: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(m.cacheFile), 0755); err != nil {
		return fmt.Errorf("error creating cache directory: %v", err)
	}

	if err := ioutil.WriteFile(m.cacheFile, data, 0644); err != nil {
		return fmt.Errorf("error writing cache: %v", err)
	}

	return nil
}

func (m *Monitor) CheckForUpdates() error {
	current, err := m.fetchCurrentVersion()
	if err != nil {
		return err
	}

	cache, err := m.loadCache()
	if err != nil {
		return err
	}

	if current != cache.LastVersion {
		path, err := m.createUpdateDirectory()
		if err != nil {
			return err
		}

		fmt.Printf("New update found! Directory created: %s\n", path)

		if err := m.saveCache(current); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	monitor := NewMonitor()
	if err := monitor.CheckForUpdates(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
