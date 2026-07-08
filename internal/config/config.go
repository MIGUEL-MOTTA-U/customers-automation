package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var DefaultConfigPath = filepath.Join(".skillsctl", "config.yaml")

type Config struct {
	Version  string           `yaml:"version"`
	Sources  []SourceConfig   `yaml:"sources"`
	Projects []ProjectProfile `yaml:"projects,omitempty"`
}

type SourceConfig struct {
	ID            string             `yaml:"id"`
	URL           string             `yaml:"url"`
	Enabled       bool               `yaml:"enabled"`
	DefaultRef    string             `yaml:"default_ref,omitempty"`
	DefaultTags   []string           `yaml:"default_tags,omitempty"`
	DefaultRoles  []string           `yaml:"default_roles,omitempty"`
	DefaultAreas  []string           `yaml:"default_areas,omitempty"`
	DefaultCases  []string           `yaml:"default_use_cases,omitempty"`
	DefaultPrio   int                `yaml:"default_priority,omitempty"`
	ProjectScopes []string           `yaml:"projects,omitempty"`
	Classifiers   []ClassifierConfig `yaml:"classifiers,omitempty"`
}

type ClassifierConfig struct {
	PathPrefix string   `yaml:"path_prefix"`
	Tags       []string `yaml:"tags,omitempty"`
	Roles      []string `yaml:"roles,omitempty"`
	UseCases   []string `yaml:"use_cases,omitempty"`
	Areas      []string `yaml:"areas,omitempty"`
	Priority   int      `yaml:"priority,omitempty"`
	Projects   []string `yaml:"projects,omitempty"`
}

type ProjectProfile struct {
	Name            string   `yaml:"name"`
	Sources         []string `yaml:"sources,omitempty"`
	Tags            []string `yaml:"tags,omitempty"`
	Roles           []string `yaml:"roles,omitempty"`
	UseCases        []string `yaml:"use_cases,omitempty"`
	Areas           []string `yaml:"areas,omitempty"`
	MaxSkills       int      `yaml:"max_skills,omitempty"`
	MinPriority     int      `yaml:"min_priority,omitempty"`
	OutputDirectory string   `yaml:"output_directory,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		Version: "1",
		Sources: []SourceConfig{},
	}
}

func Load(path string) (Config, error) {
	path = normalizeConfigPath(path)
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("no se pudo leer la configuración %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("configuración YAML inválida: %w", err)
	}
	if cfg.Version == "" {
		cfg.Version = "1"
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	path = normalizeConfigPath(path)
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("no se pudo crear directorio de config: %w", err)
	}

	content, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("no se pudo serializar la configuración: %w", err)
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("no se pudo guardar la configuración en %q: %w", path, err)
	}
	return nil
}

func normalizeConfigPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return DefaultConfigPath
	}
	return path
}

func (c Config) Validate() error {
	if len(c.Sources) == 0 {
		return nil
	}

	seenIDs := map[string]struct{}{}
	for _, src := range c.Sources {
		if strings.TrimSpace(src.ID) == "" {
			return errors.New("cada fuente debe tener un id")
		}
		if strings.TrimSpace(src.URL) == "" {
			return fmt.Errorf("la fuente %q no tiene url", src.ID)
		}
		if _, exists := seenIDs[src.ID]; exists {
			return fmt.Errorf("id de fuente duplicado: %q", src.ID)
		}
		seenIDs[src.ID] = struct{}{}
	}
	return nil
}

func (c Config) EnabledSources(filterIDs []string) []SourceConfig {
	filterSet := make(map[string]struct{})
	for _, id := range filterIDs {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			filterSet[trimmed] = struct{}{}
		}
	}

	var out []SourceConfig
	for _, src := range c.Sources {
		if !src.Enabled {
			continue
		}
		if len(filterSet) > 0 {
			if _, ok := filterSet[src.ID]; !ok {
				continue
			}
		}
		out = append(out, src)
	}
	return out
}

func (c *Config) AddSource(src SourceConfig) error {
	if strings.TrimSpace(src.ID) == "" {
		return errors.New("id de fuente requerido")
	}
	if strings.TrimSpace(src.URL) == "" {
		return errors.New("url de fuente requerida")
	}
	for _, existing := range c.Sources {
		if existing.ID == src.ID {
			return fmt.Errorf("ya existe una fuente con id %q", src.ID)
		}
		if strings.EqualFold(existing.URL, src.URL) {
			return fmt.Errorf("ya existe una fuente con url %q", src.URL)
		}
	}
	if !src.Enabled {
		src.Enabled = true
	}
	c.Sources = append(c.Sources, src)
	sort.Slice(c.Sources, func(i, j int) bool { return c.Sources[i].ID < c.Sources[j].ID })
	return nil
}

func (c Config) FindProject(name string) (ProjectProfile, bool) {
	for _, project := range c.Projects {
		if project.Name == name {
			return project, true
		}
	}
	return ProjectProfile{}, false
}
