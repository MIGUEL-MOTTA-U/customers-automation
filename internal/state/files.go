package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
	"gopkg.in/yaml.v3"
)

var (
	SelectionDir = filepath.Join(".skillsctl", "selections")
	StateDir     = filepath.Join(".skillsctl", "state")
)

func SaveSelection(project string, skills []model.SkillRef) (string, error) {
	if err := os.MkdirAll(SelectionDir, 0o755); err != nil {
		return "", fmt.Errorf("no se pudo crear directorio de selecciones: %w", err)
	}
	path := filepath.Join(SelectionDir, project+".yaml")
	payload := model.SelectionFile{
		Version:   "1",
		Project:   project,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Skills:    skills,
	}
	content, err := yaml.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("no se pudo serializar selección: %w", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", fmt.Errorf("no se pudo guardar selección %q: %w", path, err)
	}
	return path, nil
}

func LoadSelection(path string) (model.SelectionFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return model.SelectionFile{}, fmt.Errorf("no se pudo leer selección %q: %w", path, err)
	}
	var selection model.SelectionFile
	if err := yaml.Unmarshal(content, &selection); err != nil {
		return model.SelectionFile{}, fmt.Errorf("archivo de selección inválido: %w", err)
	}
	return selection, nil
}

func SaveInstallState(project string, state model.InstallState) (string, error) {
	if err := os.MkdirAll(StateDir, 0o755); err != nil {
		return "", fmt.Errorf("no se pudo crear directorio de estado: %w", err)
	}
	path := filepath.Join(StateDir, project+".json")
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", fmt.Errorf("no se pudo serializar estado de instalación: %w", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", fmt.Errorf("no se pudo escribir estado de instalación: %w", err)
	}
	return path, nil
}

func LoadInstallState(project string) (model.InstallState, string, error) {
	path := filepath.Join(StateDir, project+".json")
	content, err := os.ReadFile(path)
	if err != nil {
		return model.InstallState{}, path, fmt.Errorf("no se pudo leer estado para proyecto %q: %w", project, err)
	}
	var state model.InstallState
	if err := json.Unmarshal(content, &state); err != nil {
		return model.InstallState{}, path, fmt.Errorf("estado de instalación inválido para %q: %w", project, err)
	}
	return state, path, nil
}
