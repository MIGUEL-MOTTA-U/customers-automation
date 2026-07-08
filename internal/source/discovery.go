package source

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/config"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
)

type Discovery struct {
	github *GitHubClient
}

func NewDiscovery() *Discovery {
	return &Discovery{github: NewGitHubClient()}
}

func (d *Discovery) Discover(src config.SourceConfig) ([]model.SkillRef, error) {
	repo, err := ParseGitHubRepo(src.URL)
	if err != nil {
		return nil, fmt.Errorf("fuente %q inválida: %w", src.ID, err)
	}

	ref := src.DefaultRef
	if strings.TrimSpace(ref) == "" {
		ref, err = d.github.DefaultBranch(repo)
		if err != nil {
			return nil, fmt.Errorf("no se pudo resolver la rama por defecto de %q: %w", src.ID, err)
		}
	}

	tree, err := d.github.RepoTree(repo, ref)
	if err != nil {
		return nil, fmt.Errorf("no se pudo analizar árbol del repositorio %q: %w", src.ID, err)
	}

	candidates := candidateSkillDirs(tree)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no se encontraron skills compatibles en %q", src.ID)
	}

	skills := make([]model.SkillRef, 0, len(candidates))
	for _, dir := range candidates {
		skill := model.SkillRef{
			ID:        sanitizeID(src.ID + "-" + strings.ReplaceAll(dir, "/", "-")),
			Name:      filepath.Base(dir),
			SourceID:  src.ID,
			RepoURL:   src.URL,
			RepoPath:  dir,
			Ref:       ref,
			Tags:      append([]string{}, src.DefaultTags...),
			Roles:     append([]string{}, src.DefaultRoles...),
			UseCases:  append([]string{}, src.DefaultCases...),
			Areas:     append([]string{}, src.DefaultAreas...),
			Priority:  src.DefaultPrio,
			Projects:  append([]string{}, src.ProjectScopes...),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		applyClassifiers(&skill, src.Classifiers)
		skills = append(skills, dedupeMetadata(skill))
	}

	sort.Slice(skills, func(i, j int) bool { return skills[i].ID < skills[j].ID })
	return skills, nil
}

func candidateSkillDirs(entries []treeEntry) []string {
	dirs := map[string]struct{}{}
	for _, entry := range entries {
		if entry.Type != "blob" {
			continue
		}
		p := strings.ToLower(entry.Path)
		if strings.HasPrefix(p, ".git/") || strings.Contains(p, "/.git/") {
			continue
		}
		if strings.HasSuffix(p, "skill.yaml") ||
			strings.HasSuffix(p, "skill.yml") ||
			strings.HasSuffix(p, "skill.json") ||
			strings.HasSuffix(p, "skill.md") ||
			strings.HasSuffix(p, "readme.md") {
			dir := filepath.ToSlash(filepath.Dir(entry.Path))
			if isLikelySkillPath(dir) {
				dirs[dir] = struct{}{}
			}
		}
	}

	out := make([]string, 0, len(dirs))
	for dir := range dirs {
		out = append(out, dir)
	}
	sort.Strings(out)
	return out
}

func isLikelySkillPath(dir string) bool {
	lower := strings.ToLower(dir)
	if strings.HasPrefix(lower, "skills/") || strings.Contains(lower, "/skills/") {
		return true
	}
	segments := strings.Split(lower, "/")
	for _, segment := range segments {
		if strings.Contains(segment, "skill") {
			return true
		}
	}
	return false
}

func applyClassifiers(skill *model.SkillRef, rules []config.ClassifierConfig) {
	matchLen := -1
	for _, rule := range rules {
		prefix := filepath.ToSlash(strings.TrimSpace(rule.PathPrefix))
		if prefix == "" {
			continue
		}
		if strings.HasPrefix(skill.RepoPath, prefix) && len(prefix) > matchLen {
			matchLen = len(prefix)
			if len(rule.Tags) > 0 {
				skill.Tags = append(skill.Tags, rule.Tags...)
			}
			if len(rule.Roles) > 0 {
				skill.Roles = append(skill.Roles, rule.Roles...)
			}
			if len(rule.UseCases) > 0 {
				skill.UseCases = append(skill.UseCases, rule.UseCases...)
			}
			if len(rule.Areas) > 0 {
				skill.Areas = append(skill.Areas, rule.Areas...)
			}
			if len(rule.Projects) > 0 {
				skill.Projects = append(skill.Projects, rule.Projects...)
			}
			if rule.Priority != 0 {
				skill.Priority = rule.Priority
			}
		}
	}
}

func dedupeMetadata(skill model.SkillRef) model.SkillRef {
	skill.Tags = uniqueStrings(skill.Tags)
	skill.Roles = uniqueStrings(skill.Roles)
	skill.UseCases = uniqueStrings(skill.UseCases)
	skill.Areas = uniqueStrings(skill.Areas)
	skill.Projects = uniqueStrings(skill.Projects)
	if skill.Name == "" {
		skill.Name = filepath.Base(skill.RepoPath)
	}
	return skill
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var out []string
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func sanitizeID(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-")
	clean := replacer.Replace(input)
	for strings.Contains(clean, "--") {
		clean = strings.ReplaceAll(clean, "--", "-")
	}
	clean = strings.Trim(clean, "-")
	if clean == "" {
		return "skill"
	}
	return clean
}
