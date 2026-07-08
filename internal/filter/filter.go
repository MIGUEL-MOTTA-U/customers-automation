package filter

import (
	"sort"
	"strings"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
)

type Query struct {
	Sources     []string
	Tags        []string
	Roles       []string
	UseCases    []string
	Areas       []string
	Projects    []string
	MinPriority int
	MaxSkills   int
}

func Apply(skills []model.SkillRef, q Query) []model.SkillRef {
	if len(skills) == 0 {
		return nil
	}
	sourceSet := asSet(q.Sources)
	tagSet := asSet(q.Tags)
	roleSet := asSet(q.Roles)
	useCaseSet := asSet(q.UseCases)
	areaSet := asSet(q.Areas)
	projectSet := asSet(q.Projects)

	out := make([]model.SkillRef, 0, len(skills))
	for _, s := range skills {
		if len(sourceSet) > 0 {
			if _, ok := sourceSet[strings.ToLower(s.SourceID)]; !ok {
				continue
			}
		}
		if q.MinPriority > 0 && s.Priority < q.MinPriority {
			continue
		}
		if len(tagSet) > 0 && !intersects(tagSet, s.Tags) {
			continue
		}
		if len(roleSet) > 0 && !intersects(roleSet, s.Roles) {
			continue
		}
		if len(useCaseSet) > 0 && !intersects(useCaseSet, s.UseCases) {
			continue
		}
		if len(areaSet) > 0 && !intersects(areaSet, s.Areas) {
			continue
		}
		if len(projectSet) > 0 && !intersects(projectSet, s.Projects) {
			continue
		}
		out = append(out, s)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Priority == out[j].Priority {
			return out[i].ID < out[j].ID
		}
		return out[i].Priority > out[j].Priority
	})

	if q.MaxSkills > 0 && len(out) > q.MaxSkills {
		return out[:q.MaxSkills]
	}
	return out
}

func asSet(values []string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, value := range values {
		v := strings.ToLower(strings.TrimSpace(value))
		if v != "" {
			set[v] = struct{}{}
		}
	}
	return set
}

func intersects(expected map[string]struct{}, values []string) bool {
	for _, value := range values {
		if _, ok := expected[strings.ToLower(value)]; ok {
			return true
		}
	}
	return false
}
