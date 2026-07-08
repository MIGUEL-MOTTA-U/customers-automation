package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
)

type Catalog struct {
	Version     string           `json:"version"`
	GeneratedAt string           `json:"generated_at"`
	Skills      []model.SkillRef `json:"skills"`
}

func SaveCatalog(skills []model.SkillRef) (string, error) {
	if err := os.MkdirAll(StateDir, 0o755); err != nil {
		return "", fmt.Errorf("no se pudo crear directorio de estado: %w", err)
	}
	path := filepath.Join(StateDir, "catalog.json")
	payload := Catalog{
		Version:     "1",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Skills:      skills,
	}
	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("no se pudo serializar catálogo: %w", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", fmt.Errorf("no se pudo guardar catálogo local: %w", err)
	}
	return path, nil
}

func LoadCatalog() (Catalog, string, error) {
	path := filepath.Join(StateDir, "catalog.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return Catalog{}, path, fmt.Errorf("no se pudo leer catálogo local: %w", err)
	}
	var catalog Catalog
	if err := json.Unmarshal(content, &catalog); err != nil {
		return Catalog{}, path, fmt.Errorf("catálogo local inválido: %w", err)
	}
	return catalog, path, nil
}
