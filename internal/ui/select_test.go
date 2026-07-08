package ui_test

import (
	"io"
	"strings"
	"testing"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/ui"
)

// skillList devuelve skills de prueba para los tests de la UI.
func skillList() []model.SkillRef {
	return []model.SkillRef{
		{ID: "s1", Name: "skill-one", SourceID: "src1", RepoPath: "skills/one"},
		{ID: "s2", Name: "skill-two", SourceID: "src1", RepoPath: "skills/two"},
		{ID: "s3", Name: "skill-three", SourceID: "src2", RepoPath: "skills/three"},
	}
}

func reader(input string) io.Reader {
	return strings.NewReader(input)
}

func TestSelectSkillsInteractively_EmptyInput_ReturnsNil(t *testing.T) {
	out, err := ui.SelectSkillsInteractively(nil, reader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil for empty skills, got %v", out)
	}
}

func TestSelectSkillsInteractively_All_ReturnsAllSkills(t *testing.T) {
	skills := skillList()
	out, err := ui.SelectSkillsInteractively(skills, reader("all\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != len(skills) {
		t.Errorf("expected all %d skills, got %d", len(skills), len(out))
	}
}

func TestSelectSkillsInteractively_AllCaseInsensitive(t *testing.T) {
	skills := skillList()
	out, err := ui.SelectSkillsInteractively(skills, reader("ALL\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != len(skills) {
		t.Errorf("'ALL' should be case-insensitive: got %d", len(out))
	}
}

func TestSelectSkillsInteractively_SingleIndex(t *testing.T) {
	out, err := ui.SelectSkillsInteractively(skillList(), reader("2\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].ID != "s2" {
		t.Errorf("expected skill s2, got %v", out)
	}
}

func TestSelectSkillsInteractively_MultipleIndexes(t *testing.T) {
	out, err := ui.SelectSkillsInteractively(skillList(), reader("1,3\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(out))
	}
	if out[0].ID != "s1" || out[1].ID != "s3" {
		t.Errorf("unexpected skills: %v", out)
	}
}

func TestSelectSkillsInteractively_DuplicateIndexDeduped(t *testing.T) {
	out, err := ui.SelectSkillsInteractively(skillList(), reader("1,1,1\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Errorf("duplicate indexes should be deduped: got %d", len(out))
	}
}

func TestSelectSkillsInteractively_IndexOutOfRange_ReturnsError(t *testing.T) {
	_, err := ui.SelectSkillsInteractively(skillList(), reader("99\n"))
	if err == nil {
		t.Error("expected error for out-of-range index")
	}
}

func TestSelectSkillsInteractively_ZeroIndex_ReturnsError(t *testing.T) {
	_, err := ui.SelectSkillsInteractively(skillList(), reader("0\n"))
	if err == nil {
		t.Error("expected error for zero index (1-based indexing)")
	}
}

func TestSelectSkillsInteractively_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ui.SelectSkillsInteractively(skillList(), reader("abc\n"))
	if err == nil {
		t.Error("expected error for non-numeric input")
	}
}

func TestSelectSkillsInteractively_SpacesAroundIndexesTrimmed(t *testing.T) {
	out, err := ui.SelectSkillsInteractively(skillList(), reader(" 1 , 2 \n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2 skills with spaces trimmed, got %d", len(out))
	}
}

func TestSelectSkillsInteractively_FirstSkill(t *testing.T) {
	out, err := ui.SelectSkillsInteractively(skillList(), reader("1\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out[0].ID != "s1" {
		t.Errorf("expected s1, got %q", out[0].ID)
	}
}

func TestSelectSkillsInteractively_LastSkill(t *testing.T) {
	skills := skillList()
	out, err := ui.SelectSkillsInteractively(skills, reader("3\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out[0].ID != "s3" {
		t.Errorf("expected s3, got %q", out[0].ID)
	}
}
