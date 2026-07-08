package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/source"
)

type Layout string

const (
	LayoutDefault Layout = "default"
	LayoutClaude  Layout = "claude"
)

type Installer struct {
	github *source.GitHubClient
}

func NewInstaller() *Installer {
	return &Installer{github: source.NewGitHubClient()}
}

type Result struct {
	Records []model.InstallRecord
	Errors  []error
}

func (i *Installer) InstallSkills(skills []model.SkillRef, outputDir string, layout Layout) Result {
	if strings.TrimSpace(outputDir) == "" {
		outputDir = ".skills"
	}
	if layout == "" {
		layout = LayoutDefault
	}
	result := Result{}

	for _, skill := range skills {
		record, err := i.installSkill(skill, outputDir, layout)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", skill.ID, err))
			continue
		}
		result.Records = append(result.Records, record)
	}
	return result
}

func (i *Installer) installSkill(skill model.SkillRef, outputDir string, layout Layout) (model.InstallRecord, error) {
	repo, err := source.ParseGitHubRepo(skill.RepoURL)
	if err != nil {
		return model.InstallRecord{}, err
	}
	if strings.TrimSpace(skill.Ref) == "" {
		ref, err := i.github.DefaultBranch(repo)
		if err != nil {
			return model.InstallRecord{}, err
		}
		skill.Ref = ref
	}

	basePath := filepath.Join(outputDir, skill.SourceID, normalizeLocalName(skill.Name))
	if layout == LayoutClaude {
		basePath = filepath.Join(".claude", "skills", normalizeLocalName(skill.Name))
	}
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return model.InstallRecord{}, fmt.Errorf("no se pudo crear estructura local: %w", err)
	}

	if err := i.downloadDirectory(repo, skill.Ref, skill.RepoPath, basePath); err != nil {
		return model.InstallRecord{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	return model.InstallRecord{
		SkillID:    skill.ID,
		SourceID:   skill.SourceID,
		RepoURL:    skill.RepoURL,
		RepoPath:   skill.RepoPath,
		Ref:        skill.Ref,
		LocalPath:  basePath,
		Installed:  now,
		LastUpdate: now,
	}, nil
}

func (i *Installer) downloadDirectory(repo source.GitHubRepo, ref, remotePath, localPath string) error {
	entries, err := i.github.ListDir(repo, ref, remotePath)
	if err != nil {
		return fmt.Errorf("no se pudo listar contenido remoto %q: %w", remotePath, err)
	}
	for _, entry := range entries {
		switch entry.Type {
		case "file":
			content, err := i.github.DownloadRawFile(entry.DownloadURL)
			if err != nil {
				return fmt.Errorf("error descargando archivo %q: %w", entry.Path, err)
			}
			relative := strings.TrimPrefix(entry.Path, strings.TrimSuffix(remotePath, "/"))
			relative = strings.TrimPrefix(relative, "/")
			localFile := filepath.Join(localPath, filepath.FromSlash(relative))
			if err := os.MkdirAll(filepath.Dir(localFile), 0o755); err != nil {
				return fmt.Errorf("no se pudo crear carpeta local para %q: %w", localFile, err)
			}
			if err := os.WriteFile(localFile, content, 0o644); err != nil {
				return fmt.Errorf("no se pudo guardar archivo local %q: %w", localFile, err)
			}
		case "dir":
			subRemote := entry.Path
			subLocal := filepath.Join(localPath, filepath.Base(subRemote))
			if err := os.MkdirAll(subLocal, 0o755); err != nil {
				return fmt.Errorf("no se pudo crear subdirectorio %q: %w", subLocal, err)
			}
			if err := i.downloadDirectory(repo, ref, subRemote, subLocal); err != nil {
				return err
			}
		}
	}
	return nil
}

func normalizeLocalName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	r := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-")
	n = r.Replace(n)
	for strings.Contains(n, "--") {
		n = strings.ReplaceAll(n, "--", "-")
	}
	n = strings.Trim(n, "-")
	if n == "" {
		return "skill"
	}
	return n
}
