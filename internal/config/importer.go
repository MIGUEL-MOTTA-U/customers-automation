package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	markdownLinkPattern = regexp.MustCompile(`https://github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+`)
)

func BuildConfigFromLinksFile(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("no se pudo leer archivo de enlaces %q: %w", path, err)
	}

	text := string(content)
	links := markdownLinkPattern.FindAllString(text, -1)
	seen := map[string]struct{}{}
	cfg := DefaultConfig()

	for _, link := range links {
		if _, exists := seen[link]; exists {
			continue
		}
		seen[link] = struct{}{}
		id := sourceIDFromURL(link)
		cfg.Sources = append(cfg.Sources, SourceConfig{
			ID:      id,
			URL:     link,
			Enabled: true,
		})
	}
	if len(cfg.Sources) == 0 {
		return Config{}, fmt.Errorf("no se encontraron enlaces válidos de GitHub en %q", path)
	}
	return cfg, nil
}

func sourceIDFromURL(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimPrefix(url, "https://github.com/")
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "source"
	}
	id := parts[0] + "-" + parts[1]
	return strings.ToLower(strings.ReplaceAll(id, "_", "-"))
}
