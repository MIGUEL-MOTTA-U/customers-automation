package state_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/state"
)

// withTempStateDirs reemplaza SelectionDir y StateDir por directorios temporales
// y devuelve una función restore que los restablece al finalizar.
func withTempStateDirs(t *testing.T) func() {
	t.Helper()
	origSel := state.SelectionDir
	origSt := state.StateDir
	dir := t.TempDir()
	state.SelectionDir = filepath.Join(dir, "selections")
	state.StateDir = filepath.Join(dir, "state")
	return func() {
		state.SelectionDir = origSel
		state.StateDir = origSt
	}
}

// ── Catalog ───────────────────────────────────────────────────────────────────

func TestSaveCatalog_CreatesFile(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	skills := []model.SkillRef{{ID: "skill-a", Name: "Skill A", SourceID: "src1"}}
	path, err := state.SaveCatalog(skills)
	if err != nil {
		t.Fatalf("SaveCatalog: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("catalog file should exist at %q: %v", path, err)
	}
}

func TestLoadCatalog_RoundTrip(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	skills := []model.SkillRef{
		{ID: "a", Name: "Skill A", SourceID: "src1"},
		{ID: "b", Name: "Skill B", SourceID: "src2"},
	}
	if _, err := state.SaveCatalog(skills); err != nil {
		t.Fatalf("SaveCatalog: %v", err)
	}

	catalog, _, err := state.LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	if len(catalog.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(catalog.Skills))
	}
	if catalog.Skills[0].ID != "a" || catalog.Skills[1].ID != "b" {
		t.Errorf("unexpected skills after round-trip: %v", catalog.Skills)
	}
}

func TestLoadCatalog_SetsVersionAndTimestamp(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	_, _ = state.SaveCatalog([]model.SkillRef{{ID: "x"}})
	catalog, _, err := state.LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Version != "1" {
		t.Errorf("expected version 1, got %q", catalog.Version)
	}
	if catalog.GeneratedAt == "" {
		t.Error("GeneratedAt should be set")
	}
}

func TestLoadCatalog_MissingFile(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	_, _, err := state.LoadCatalog()
	if err == nil {
		t.Error("expected error when catalog file does not exist")
	}
}

// ── Selection ─────────────────────────────────────────────────────────────────

func TestSaveSelection_CreatesFile(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	skills := []model.SkillRef{{ID: "s1", Name: "Skill 1"}}
	path, err := state.SaveSelection("my-project", skills)
	if err != nil {
		t.Fatalf("SaveSelection: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("selection file should exist: %v", err)
	}
}

func TestLoadSelection_RoundTrip(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	skills := []model.SkillRef{
		{ID: "s1", Name: "Skill 1", SourceID: "src1"},
	}
	path, err := state.SaveSelection("proj-x", skills)
	if err != nil {
		t.Fatalf("SaveSelection: %v", err)
	}

	sel, err := state.LoadSelection(path)
	if err != nil {
		t.Fatalf("LoadSelection: %v", err)
	}
	if sel.Project != "proj-x" {
		t.Errorf("expected project proj-x, got %q", sel.Project)
	}
	if len(sel.Skills) != 1 || sel.Skills[0].ID != "s1" {
		t.Errorf("unexpected skills in selection: %v", sel.Skills)
	}
}

func TestSaveSelection_SetsCreatedAt(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	path, _ := state.SaveSelection("proj", []model.SkillRef{{ID: "x"}})
	sel, err := state.LoadSelection(path)
	if err != nil {
		t.Fatal(err)
	}
	if sel.CreatedAt == "" {
		t.Error("CreatedAt should be set")
	}
}

func TestLoadSelection_NonExistentFile(t *testing.T) {
	_, err := state.LoadSelection("/this/does/not/exist.yaml")
	if err == nil {
		t.Error("expected error for missing selection file")
	}
}

// ── InstallState ─────────────────────────────────────────────────────────────

func TestSaveInstallState_CreatesFile(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	s := model.InstallState{
		Version: "1",
		Project: "test-proj",
		Records: []model.InstallRecord{
			{SkillID: "sk1", LocalPath: ".skills/sk1", SourceID: "src1"},
		},
	}
	path, err := state.SaveInstallState("test-proj", s)
	if err != nil {
		t.Fatalf("SaveInstallState: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("install state file should exist: %v", err)
	}
}

func TestLoadInstallState_RoundTrip(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	original := model.InstallState{
		Version: "1",
		Project: "round-trip",
		Records: []model.InstallRecord{
			{SkillID: "sk1", LocalPath: ".skills/sk1", Ref: "main"},
			{SkillID: "sk2", LocalPath: ".skills/sk2", Ref: "main"},
		},
	}
	if _, err := state.SaveInstallState("round-trip", original); err != nil {
		t.Fatalf("SaveInstallState: %v", err)
	}

	loaded, _, err := state.LoadInstallState("round-trip")
	if err != nil {
		t.Fatalf("LoadInstallState: %v", err)
	}
	if loaded.Project != "round-trip" {
		t.Errorf("unexpected project: %q", loaded.Project)
	}
	if len(loaded.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(loaded.Records))
	}
	if loaded.Records[0].SkillID != "sk1" {
		t.Errorf("unexpected skill ID: %q", loaded.Records[0].SkillID)
	}
}

func TestLoadInstallState_MissingProject(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	_, _, err := state.LoadInstallState("nonexistent-project")
	if err == nil {
		t.Error("expected error for missing project state")
	}
}

func TestSaveInstallState_MultipleProjects(t *testing.T) {
	restore := withTempStateDirs(t)
	defer restore()

	for _, proj := range []string{"proj-a", "proj-b", "proj-c"} {
		s := model.InstallState{Version: "1", Project: proj}
		if _, err := state.SaveInstallState(proj, s); err != nil {
			t.Fatalf("SaveInstallState(%s): %v", proj, err)
		}
	}

	for _, proj := range []string{"proj-a", "proj-b", "proj-c"} {
		loaded, _, err := state.LoadInstallState(proj)
		if err != nil {
			t.Errorf("LoadInstallState(%s): %v", proj, err)
		}
		if loaded.Project != proj {
			t.Errorf("expected project %q, got %q", proj, loaded.Project)
		}
	}
}
