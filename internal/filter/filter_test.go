package filter_test

import (
	"testing"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/filter"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
)

// skillSet devuelve un conjunto de skills de prueba con distintas taxonomías.
func skillSet() []model.SkillRef {
	return []model.SkillRef{
		{
			ID:       "a",
			SourceID: "src1",
			Tags:     []string{"go", "cli"},
			Roles:    []string{"backend"},
			UseCases: []string{"dev-env"},
			Areas:    []string{"automation"},
			Priority: 5,
			Projects: []string{"proj-a"},
		},
		{
			ID:       "b",
			SourceID: "src1",
			Tags:     []string{"security"},
			Roles:    []string{"security-engineer"},
			UseCases: []string{"audit"},
			Areas:    []string{"cybersecurity"},
			Priority: 9,
			Projects: []string{"proj-b"},
		},
		{
			ID:       "c",
			SourceID: "src2",
			Tags:     []string{"go"},
			Roles:    []string{"backend"},
			UseCases: []string{"ci"},
			Areas:    []string{"devops"},
			Priority: 3,
			Projects: []string{"proj-a"},
		},
	}
}

func TestApply_NoFilter_ReturnsSortedByPriorityDesc(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{})
	if len(out) != 3 {
		t.Fatalf("expected 3 results, got %d", len(out))
	}
	// prioridad desc: b(9), a(5), c(3)
	if out[0].ID != "b" || out[1].ID != "a" || out[2].ID != "c" {
		t.Errorf("orden inesperado: %v %v %v", out[0].ID, out[1].ID, out[2].ID)
	}
}

func TestApply_EmptyInput_ReturnsNil(t *testing.T) {
	out := filter.Apply(nil, filter.Query{})
	if out != nil {
		t.Errorf("expected nil for empty input, got %v", out)
	}
}

func TestApply_FilterBySource_SingleSource(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{Sources: []string{"src2"}})
	if len(out) != 1 || out[0].ID != "c" {
		t.Errorf("expected only skill c from src2, got %v", out)
	}
}

func TestApply_FilterBySource_CaseInsensitive(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{Sources: []string{"SRC1"}})
	if len(out) != 2 {
		t.Errorf("expected 2 skills from src1, got %d", len(out))
	}
}

func TestApply_FilterByTag_SingleMatch(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{Tags: []string{"security"}})
	if len(out) != 1 || out[0].ID != "b" {
		t.Errorf("expected skill b for tag security, got %v", out)
	}
}

func TestApply_FilterByTag_MultipleMatches(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{Tags: []string{"go"}})
	if len(out) != 2 {
		t.Errorf("expected 2 skills with tag go, got %d", len(out))
	}
}

func TestApply_FilterByTag_CaseInsensitive(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{Tags: []string{"GO"}})
	if len(out) != 2 {
		t.Errorf("tag filter should be case-insensitive, got %d", len(out))
	}
}

func TestApply_FilterByRole(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{Roles: []string{"security-engineer"}})
	if len(out) != 1 || out[0].ID != "b" {
		t.Errorf("expected skill b for role security-engineer, got %v", out)
	}
}

func TestApply_FilterByUseCase(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{UseCases: []string{"audit"}})
	if len(out) != 1 || out[0].ID != "b" {
		t.Errorf("expected skill b for use-case audit, got %v", out)
	}
}

func TestApply_FilterByArea(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{Areas: []string{"devops"}})
	if len(out) != 1 || out[0].ID != "c" {
		t.Errorf("expected skill c for area devops, got %v", out)
	}
}

func TestApply_FilterByProject(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{Projects: []string{"proj-a"}})
	if len(out) != 2 {
		t.Errorf("expected 2 skills in proj-a, got %d", len(out))
	}
}

func TestApply_FilterByProject_NoMatch(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{Projects: []string{"nonexistent"}})
	if len(out) != 0 {
		t.Errorf("expected 0 results for unknown project, got %d", len(out))
	}
}

func TestApply_FilterByMinPriority_IncludesEqual(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{MinPriority: 5})
	if len(out) != 2 {
		t.Errorf("expected 2 skills with priority >= 5, got %d", len(out))
	}
}

func TestApply_FilterByMinPriority_ExcludesBelow(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{MinPriority: 10})
	if len(out) != 0 {
		t.Errorf("expected 0 skills with priority >= 10, got %d", len(out))
	}
}

func TestApply_MaxSkills_LimitsResults(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{MaxSkills: 1})
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	// El primero debe ser el de mayor prioridad (b=9)
	if out[0].ID != "b" {
		t.Errorf("expected highest priority skill b, got %s", out[0].ID)
	}
}

func TestApply_MaxSkills_ZeroMeansUnlimited(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{MaxSkills: 0})
	if len(out) != 3 {
		t.Errorf("MaxSkills=0 should return all, got %d", len(out))
	}
}

func TestApply_CombinedFilters(t *testing.T) {
	out := filter.Apply(skillSet(), filter.Query{
		Tags:        []string{"go"},
		Roles:       []string{"backend"},
		MinPriority: 5,
	})
	if len(out) != 1 || out[0].ID != "a" {
		t.Errorf("expected only skill a for combined filters, got %v", out)
	}
}

func TestApply_TiePriorityOrderedByID(t *testing.T) {
	skills := []model.SkillRef{
		{ID: "z", Priority: 5},
		{ID: "a", Priority: 5},
		{ID: "m", Priority: 5},
	}
	out := filter.Apply(skills, filter.Query{})
	if out[0].ID != "a" || out[1].ID != "m" || out[2].ID != "z" {
		t.Errorf("tie in priority should sort by ID: got %v %v %v", out[0].ID, out[1].ID, out[2].ID)
	}
}
