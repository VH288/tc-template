package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

type SwaggerConfig struct {
	URLs        []SwaggerURL `json:"urls"`
	PrimaryName string       `json:"urls.primaryName,omitempty"`
}

type SwaggerURL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type openAPISpec struct {
	Info struct {
		Title string `json:"title"`
	} `json:"info"`
}

func findRootDir() (string, error) {
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "documentation", "api")); err == nil {
			return cwd, nil
		}
	}

	_, filename, _, ok := runtime.Caller(0)
	if ok {
		parent := filepath.Dir(filepath.Dir(filename))
		if _, err := os.Stat(filepath.Join(parent, "documentation", "api")); err == nil {
			return parent, nil
		}
	}

	return "", fmt.Errorf("unable to locate documentation/api directory")
}

func extractTitleFromJSON(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var spec openAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return "", err
	}
	return strings.TrimSpace(spec.Info.Title), nil
}

func extractTitleFromYAML(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	titleRe := regexp.MustCompile(`^\s*title:\s*["']?(.*?)["']?\s*$`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := titleRe.FindStringSubmatch(line)
		if len(matches) > 1 && strings.TrimSpace(matches[1]) != "" {
			return strings.TrimSpace(matches[1]), nil
		}
	}
	return "", scanner.Err()
}

func formatFallbackTitle(dirName string) string {
	clean := strings.ReplaceAll(dirName, "-", " ")
	clean = strings.ReplaceAll(clean, "_", " ")
	words := strings.Fields(clean)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ") + " API"
}

func main() {
	rootDir, err := findRootDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	apiDir := filepath.Join(rootDir, "documentation", "api")
	entries, err := os.ReadDir(apiDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read api dir: %v\n", err)
		os.Exit(1)
	}

	var urls []SwaggerURL

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		modelName := entry.Name()
		modelDir := filepath.Join(apiDir, modelName)

		jsonPath := filepath.Join(modelDir, "openapi.json")
		yamlPath := filepath.Join(modelDir, "openapi.yaml")
		ymlPath := filepath.Join(modelDir, "openapi.yml")

		var title string
		var specRelURL string

		if _, err := os.Stat(jsonPath); err == nil {
			specRelURL = fmt.Sprintf("/docs/%s/openapi.json", modelName)
			t, err := extractTitleFromJSON(jsonPath)
			if err == nil && t != "" {
				title = t
			}
		} else if _, err := os.Stat(yamlPath); err == nil {
			specRelURL = fmt.Sprintf("/docs/%s/openapi.yaml", modelName)
			t, err := extractTitleFromYAML(yamlPath)
			if err == nil && t != "" {
				title = t
			}
		} else if _, err := os.Stat(ymlPath); err == nil {
			specRelURL = fmt.Sprintf("/docs/%s/openapi.yml", modelName)
			t, err := extractTitleFromYAML(ymlPath)
			if err == nil && t != "" {
				title = t
			}
		}

		if specRelURL == "" {
			continue
		}

		if title == "" {
			title = formatFallbackTitle(modelName)
		}

		urls = append(urls, SwaggerURL{
			Name: title,
			URL:  specRelURL,
		})
	}

	if len(urls) == 0 {
		fmt.Println("No OpenAPI specifications found.")
		return
	}

	// Sort URLs: "Sample API" first if present, others alphabetically by Name
	sort.SliceStable(urls, func(i, j int) bool {
		if strings.EqualFold(urls[i].Name, "Sample API") {
			return true
		}
		if strings.EqualFold(urls[j].Name, "Sample API") {
			return false
		}
		return urls[i].Name < urls[j].Name
	})

	primaryName := urls[0].Name
	for _, u := range urls {
		if strings.EqualFold(u.Name, "Sample API") {
			primaryName = u.Name
			break
		}
	}

	config := SwaggerConfig{
		URLs:        urls,
		PrimaryName: primaryName,
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal config: %v\n", err)
		os.Exit(1)
	}
	configData = append(configData, '\n')

	targetPath := filepath.Join(apiDir, "swagger-config.json")
	if err := os.WriteFile(targetPath, configData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", targetPath, err)
		os.Exit(1)
	}

	fmt.Printf("Discovered %d API specifications:\n", len(urls))
	for _, u := range urls {
		fmt.Printf("  - %s (%s)\n", u.Name, u.URL)
	}
	fmt.Printf("Updated %s successfully (primary: %s)\n", targetPath, primaryName)
}
