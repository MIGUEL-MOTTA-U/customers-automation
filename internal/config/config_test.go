package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/config"
)

// ── DefaultConfig ─────────────────────────────────────────────────────────────

func TestDefaultConfig_Version(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.Version != "1" {
		t.Errorf("expected version 1, got %q", cfg.Version)
	}
}

func TestDefaultConfig_EmptySourcesSlice(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.Sources == nil {
		t.Error("Sources should be empty slice, not nil")
	}
}

// ── Config.Validate ───────────────────────────────────────────────────────────

func TestValidate_EmptySourcesIsValid(t *testing.T) {
	cfg := config.DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("empty sources should be valid: %v", err)
	}
}

func TestValidate_ValidSingleSource(t *testing.T) {
	cfg := config.Config{
		Sources: []config.SourceConfig{
			{ID: "a", URL: "https://github.com/a/a"},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_MissingID(t *testing.T) {
	cfg := config.Config{
		Sources: []config.SourceConfig{
			{ID: "", URL: "https://github.com/a/a"},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing ID")
	}
}

func TestValidate_MissingURL(t *testing.T) {
	cfg := config.Config{
		Sources: []config.SourceConfig{
			{ID: "a", URL: ""},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestValidate_DuplicateIDs(t *testing.T) {
	cfg := config.Config{
		Sources: []config.SourceConfig{
			{ID: "dup", URL: "https://github.com/a/a"},
			{ID: "dup", URL: "https://github.com/b/b"},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected duplicate ID error")
	}
}

// ── Config.AddSource ──────────────────────────────────────────────────────────

func TestAddSource_Success(t *testing.T) {
	cfg := config.DefaultConfig()
	err := cfg.AddSource(config.SourceConfig{ID: "a", URL: "https://github.com/a/a", Enabled: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Sources) != 1 {
		t.Errorf("expected 1 source, got %d", len(cfg.Sources))
	}
}

func TestAddSource_EnabledForcedTrue(t *testing.T) {
	cfg := config.DefaultConfig()
	_ = cfg.AddSource(config.SourceConfig{ID: "a", URL: "https://github.com/a/a", Enabled: false})
	if !cfg.Sources[0].Enabled {
		t.Error("AddSource should force Enabled=true")
	}
}

func TestAddSource_MissingID(t *testing.T) {
	cfg := config.DefaultConfig()
	err := cfg.AddSource(config.SourceConfig{ID: "", URL: "https://github.com/a/a"})
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestAddSource_MissingURL(t *testing.T) {
	cfg := config.DefaultConfig()
	err := cfg.AddSource(config.SourceConfig{ID: "a", URL: ""})
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestAddSource_DuplicateID(t *testing.T) {
	cfg := config.DefaultConfig()
	_ = cfg.AddSource(config.SourceConfig{ID: "a", URL: "https://github.com/a/a", Enabled: true})
	err := cfg.AddSource(config.SourceConfig{ID: "a", URL: "https://github.com/b/b", Enabled: true})
	if err == nil {
		t.Error("expected duplicate ID error")
	}
}

func TestAddSource_DuplicateURL(t *testing.T) {
	cfg := config.DefaultConfig()
	_ = cfg.AddSource(config.SourceConfig{ID: "a", URL: "https://github.com/a/a", Enabled: true})
	err := cfg.AddSource(config.SourceConfig{ID: "b", URL: "https://github.com/A/A", Enabled: true}) // case-insensitive
	if err == nil {
		t.Error("expected duplicate URL error (case-insensitive)")
	}
}

func TestAddSource_SortsByID(t *testing.T) {
	cfg := config.DefaultConfig()
	_ = cfg.AddSource(config.SourceConfig{ID: "z", URL: "https://github.com/z/z", Enabled: true})
	_ = cfg.AddSource(config.SourceConfig{ID: "a", URL: "https://github.com/a/a", Enabled: true})
	_ = cfg.AddSource(config.SourceConfig{ID: "m", URL: "https://github.com/m/m", Enabled: true})
	if cfg.Sources[0].ID != "a" || cfg.Sources[1].ID != "m" || cfg.Sources[2].ID != "z" {
		t.Errorf("sources not sorted: %v", cfg.Sources)
	}
}

// ── Config.EnabledSources ─────────────────────────────────────────────────────

func TestEnabledSources_OnlyEnabled(t *testing.T) {
	cfg := config.Config{
		Sources: []config.SourceConfig{
			{ID: "a", URL: "https://github.com/a/a", Enabled: true},
			{ID: "b", URL: "https://github.com/b/b", Enabled: false},
		},
	}
	out := cfg.EnabledSources(nil)
	if len(out) != 1 || out[0].ID != "a" {
		t.Errorf("expected only enabled source: %v", out)
	}
}

func TestEnabledSources_FilterByID(t *testing.T) {
	cfg := config.Config{
		Sources: []config.SourceConfig{
			{ID: "a", URL: "https://github.com/a/a", Enabled: true},
			{ID: "b", URL: "https://github.com/b/b", Enabled: true},
		},
	}
	out := cfg.EnabledSources([]string{"b"})
	if len(out) != 1 || out[0].ID != "b" {
		t.Errorf("expected only source b: %v", out)
	}
}

func TestEnabledSources_NoFilterReturnsAll(t *testing.T) {
	cfg := config.Config{
		Sources: []config.SourceConfig{
			{ID: "a", URL: "https://github.com/a/a", Enabled: true},
			{ID: "b", URL: "https://github.com/b/b", Enabled: true},
		},
	}
	out := cfg.EnabledSources(nil)
	if len(out) != 2 {
		t.Errorf("expected both sources, got %d", len(out))
	}
}

func TestEnabledSources_UnknownIDReturnsEmpty(t *testing.T) {
	cfg := config.Config{
		Sources: []config.SourceConfig{
			{ID: "a", URL: "https://github.com/a/a", Enabled: true},
		},
	}
	out := cfg.EnabledSources([]string{"nonexistent"})
	if len(out) != 0 {
		t.Errorf("expected 0 results for unknown ID, got %d", len(out))
	}
}

// ── Save / Load ───────────────────────────────────────────────────────────────

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := config.DefaultConfig()
	_ = cfg.AddSource(config.SourceConfig{
		ID:          "test-source",
		URL:         "https://github.com/test/repo",
		Enabled:     true,
		DefaultTags: []string{"go", "cli"},
	})

	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded.Sources) != 1 {
		t.Fatalf("expected 1 source after round-trip, got %d", len(loaded.Sources))
	}
	if loaded.Sources[0].ID != "test-source" {
		t.Errorf("unexpected source ID: %q", loaded.Sources[0].ID)
	}
	if len(loaded.Sources[0].DefaultTags) != 2 {
		t.Errorf("expected 2 tags after round-trip, got %v", loaded.Sources[0].DefaultTags)
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := config.Load("/this/path/does/not/exist/config.yaml")
	if err == nil {
		t.Error("expected error loading nonexistent file")
	}
}

func TestSave_CreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "config.yaml")

	if err := config.Save(path, config.DefaultConfig()); err != nil {
		t.Fatalf("Save should create parent dirs: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

// ── Config.FindProject ────────────────────────────────────────────────────────

func TestFindProject_Found(t *testing.T) {
	cfg := config.Config{
		Projects: []config.ProjectProfile{
			{Name: "proj-a"},
			{Name: "proj-b"},
		},
	}
	p, ok := cfg.FindProject("proj-a")
	if !ok {
		t.Error("expected to find proj-a")
	}
	if p.Name != "proj-a" {
		t.Errorf("unexpected project: %v", p)
	}
}

func TestFindProject_NotFound(t *testing.T) {
	cfg := config.Config{
		Projects: []config.ProjectProfile{{Name: "proj-a"}},
	}
	_, ok := cfg.FindProject("nonexistent")
	if ok {
		t.Error("expected not found for nonexistent project")
	}
}

// ── BuildConfigFromLinksFile ─────────────────────────────────────────────────

func TestBuildConfigFromLinksFile_ValidMarkdown(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "sources.md")
	content := "# Sources\n- https://github.com/owner1/repo1\n- https://github.com/owner2/repo2\n"
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.BuildConfigFromLinksFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(cfg.Sources))
	}
}

func TestBuildConfigFromLinksFile_DeduplicatesLinks(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "links.md")
	_ = os.WriteFile(f, []byte("https://github.com/a/b\nhttps://github.com/a/b\n"), 0o644)

	cfg, err := config.BuildConfigFromLinksFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sources) != 1 {
		t.Errorf("expected 1 deduplicated source, got %d", len(cfg.Sources))
	}
}

func TestBuildConfigFromLinksFile_NoValidLinks(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "empty.md")
	_ = os.WriteFile(f, []byte("no github links here\n"), 0o644)

	_, err := config.BuildConfigFromLinksFile(f)
	if err == nil {
		t.Error("expected error when no GitHub links found")
	}
}

func TestBuildConfigFromLinksFile_NonExistentFile(t *testing.T) {
	_, err := config.BuildConfigFromLinksFile("/nonexistent/file.md")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestBuildConfigFromLinksFile_SourceIDFromURL(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "links.txt")
	_ = os.WriteFile(f, []byte("https://github.com/MyOrg/My_Repo\n"), 0o644)

	cfg, err := config.BuildConfigFromLinksFile(f)
	if err != nil {
		t.Fatal(err)
	}
	// ID should be lowercase hyphenated: myorg-my-repo
	if cfg.Sources[0].ID != "myorg-my-repo" {
		t.Errorf("unexpected source ID: %q", cfg.Sources[0].ID)
	}
}
