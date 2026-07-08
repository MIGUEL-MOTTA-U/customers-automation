package source

import (
	"testing"

	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/config"
	"github.com/MIGUEL-MOTTA-U/customers-automation/internal/model"
)

// ── ParseGitHubRepo ──────────────────────────────────────────────────────────

func TestParseGitHubRepo_Valid(t *testing.T) {
	repo, err := ParseGitHubRepo("https://github.com/owner/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Owner != "owner" || repo.Repo != "repo" {
		t.Errorf("unexpected parsed repo: %+v", repo)
	}
}

func TestParseGitHubRepo_TrailingSlash(t *testing.T) {
	repo, err := ParseGitHubRepo("https://github.com/owner/repo/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Owner != "owner" || repo.Repo != "repo" {
		t.Errorf("trailing slash not handled: %+v", repo)
	}
}

func TestParseGitHubRepo_UnsupportedHost(t *testing.T) {
	_, err := ParseGitHubRepo("https://gitlab.com/owner/repo")
	if err == nil {
		t.Error("expected error for gitlab.com host")
	}
}

func TestParseGitHubRepo_MissingRepoSegment(t *testing.T) {
	_, err := ParseGitHubRepo("https://github.com/onlyowner")
	if err == nil {
		t.Error("expected error when repo path segment is missing")
	}
}

func TestParseGitHubRepo_EmptyURL(t *testing.T) {
	_, err := ParseGitHubRepo("")
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestParseGitHubRepo_InvalidURL(t *testing.T) {
	_, err := ParseGitHubRepo("not-a-url")
	if err == nil {
		t.Error("expected error for invalid URL without github.com host")
	}
}

// ── sanitizeID ───────────────────────────────────────────────────────────────

func TestSanitizeID_SlashBecomesHyphen(t *testing.T) {
	got := sanitizeID("hello/world")
	if got != "hello-world" {
		t.Errorf("got %q, want %q", got, "hello-world")
	}
}

func TestSanitizeID_SpaceBecomesHyphen(t *testing.T) {
	got := sanitizeID("Hello World")
	if got != "hello-world" {
		t.Errorf("got %q, want %q", got, "hello-world")
	}
}

func TestSanitizeID_ConsecutiveHyphensCollapsed(t *testing.T) {
	got := sanitizeID("a--b")
	if got != "a-b" {
		t.Errorf("got %q, want %q", got, "a-b")
	}
}

func TestSanitizeID_LeadingTrailingHyphensTrimmed(t *testing.T) {
	got := sanitizeID("-trim-")
	if got != "trim" {
		t.Errorf("got %q, want %q", got, "trim")
	}
}

func TestSanitizeID_EmptyFallback(t *testing.T) {
	got := sanitizeID("")
	if got != "skill" {
		t.Errorf("empty input should return %q, got %q", "skill", got)
	}
}

func TestSanitizeID_UnderscoreAndDotConverted(t *testing.T) {
	got := sanitizeID("a_b.c")
	if got != "a-b-c" {
		t.Errorf("got %q, want %q", got, "a-b-c")
	}
}

func TestSanitizeID_AllLowercase(t *testing.T) {
	got := sanitizeID("MySkill")
	if got != "myskill" {
		t.Errorf("got %q, want lowercase %q", got, "myskill")
	}
}

// ── uniqueStrings ─────────────────────────────────────────────────────────────

func TestUniqueStrings_DeduplicatesCaseInsensitive(t *testing.T) {
	in := []string{"go", "Go", " go ", "cli", "CLI"}
	out := uniqueStrings(in)
	if len(out) != 2 {
		t.Errorf("expected 2 unique values, got %d: %v", len(out), out)
	}
}

func TestUniqueStrings_EmptyInput(t *testing.T) {
	out := uniqueStrings(nil)
	if len(out) != 0 {
		t.Errorf("expected empty output, got %v", out)
	}
}

func TestUniqueStrings_SkipsBlankEntries(t *testing.T) {
	out := uniqueStrings([]string{"a", "", "  "})
	if len(out) != 1 || out[0] != "a" {
		t.Errorf("should skip blank entries, got %v", out)
	}
}

func TestUniqueStrings_Sorted(t *testing.T) {
	out := uniqueStrings([]string{"z", "a", "m"})
	if out[0] != "a" || out[1] != "m" || out[2] != "z" {
		t.Errorf("output should be sorted, got %v", out)
	}
}

// ── isLikelySkillPath ─────────────────────────────────────────────────────────

func TestIsLikelySkillPath_SkillsPrefix(t *testing.T) {
	if !isLikelySkillPath("skills/code-review") {
		t.Error("skills/ prefix should match")
	}
}

func TestIsLikelySkillPath_NestedSkills(t *testing.T) {
	if !isLikelySkillPath("a/skills/b") {
		t.Error("/skills/ segment should match")
	}
}

func TestIsLikelySkillPath_SegmentContainsSkill(t *testing.T) {
	if !isLikelySkillPath("myskills/subfolder") {
		t.Error("segment containing 'skill' should match")
	}
}

func TestIsLikelySkillPath_UnrelatedPath(t *testing.T) {
	if isLikelySkillPath("docs/api") {
		t.Error("docs/api should NOT match")
	}
}

func TestIsLikelySkillPath_EmptyPath(t *testing.T) {
	if isLikelySkillPath("") {
		t.Error("empty path should NOT match")
	}
}

// ── candidateSkillDirs ────────────────────────────────────────────────────────

func TestCandidateSkillDirs_SkillYaml(t *testing.T) {
	entries := []treeEntry{
		{Path: "skills/code-review/skill.yaml", Type: "blob"},
		{Path: "docs/readme.md", Type: "blob"},
	}
	dirs := candidateSkillDirs(entries)
	if len(dirs) != 1 || dirs[0] != "skills/code-review" {
		t.Errorf("unexpected dirs: %v", dirs)
	}
}

func TestCandidateSkillDirs_IgnoresGitDir(t *testing.T) {
	entries := []treeEntry{
		{Path: ".git/skills/skill.yaml", Type: "blob"},
	}
	dirs := candidateSkillDirs(entries)
	if len(dirs) != 0 {
		t.Errorf("should ignore .git paths: %v", dirs)
	}
}

func TestCandidateSkillDirs_ReadmeInSkillPath(t *testing.T) {
	entries := []treeEntry{
		{Path: "skills/my-skill/README.md", Type: "blob"},
	}
	dirs := candidateSkillDirs(entries)
	if len(dirs) != 1 {
		t.Errorf("README.md in skills path should be a candidate: %v", dirs)
	}
}

func TestCandidateSkillDirs_IgnoresTreeEntries(t *testing.T) {
	// Type "tree" should NOT be considered (only "blob")
	entries := []treeEntry{
		{Path: "skills/my-skill", Type: "tree"},
	}
	dirs := candidateSkillDirs(entries)
	if len(dirs) != 0 {
		t.Errorf("tree entries should be ignored: %v", dirs)
	}
}

func TestCandidateSkillDirs_Deduplicates(t *testing.T) {
	entries := []treeEntry{
		{Path: "skills/review/skill.yaml", Type: "blob"},
		{Path: "skills/review/README.md", Type: "blob"},
	}
	dirs := candidateSkillDirs(entries)
	if len(dirs) != 1 {
		t.Errorf("same directory should appear only once, got %v", dirs)
	}
}

func TestCandidateSkillDirs_MultipleSkills(t *testing.T) {
	entries := []treeEntry{
		{Path: "skills/a/skill.yaml", Type: "blob"},
		{Path: "skills/b/skill.md", Type: "blob"},
		{Path: "skills/c/skill.yml", Type: "blob"},
	}
	dirs := candidateSkillDirs(entries)
	if len(dirs) != 3 {
		t.Errorf("expected 3 skill dirs, got %d: %v", len(dirs), dirs)
	}
}

// ── dedupeMetadata ────────────────────────────────────────────────────────────

func TestDedupeMetadata_SetsNameFromRepoPath(t *testing.T) {
	skill := model.SkillRef{RepoPath: "skills/test-skill"}
	out := dedupeMetadata(skill)
	if out.Name != "test-skill" {
		t.Errorf("expected Name from RepoPath base, got %q", out.Name)
	}
}

func TestDedupeMetadata_KeepsExistingName(t *testing.T) {
	skill := model.SkillRef{Name: "custom-name", RepoPath: "skills/other"}
	out := dedupeMetadata(skill)
	if out.Name != "custom-name" {
		t.Errorf("should keep existing Name, got %q", out.Name)
	}
}

func TestDedupeMetadata_DeduplicatesTags(t *testing.T) {
	skill := model.SkillRef{
		Name: "s",
		Tags: []string{"go", "Go", "cli"},
	}
	out := dedupeMetadata(skill)
	if len(out.Tags) != 2 {
		t.Errorf("expected 2 unique tags, got %v", out.Tags)
	}
}

// ── applyClassifiers ─────────────────────────────────────────────────────────

func TestApplyClassifiers_LongestPrefixWins(t *testing.T) {
	skill := &model.SkillRef{RepoPath: "skills/security/audit"}
	rules := []config.ClassifierConfig{
		{PathPrefix: "skills", Tags: []string{"generic"}, Priority: 1},
		{PathPrefix: "skills/security", Tags: []string{"security"}, Priority: 9},
	}
	applyClassifiers(skill, rules)
	// El prefijo más largo (skills/security) determina la prioridad final
	if skill.Priority != 9 {
		t.Errorf("longest prefix should win priority: got %d, want 9", skill.Priority)
	}
}

func TestApplyClassifiers_TagsAccumulate(t *testing.T) {
	skill := &model.SkillRef{RepoPath: "skills/security/x", Tags: []string{"existing"}}
	rules := []config.ClassifierConfig{
		{PathPrefix: "skills/security", Tags: []string{"security"}, Priority: 9},
	}
	applyClassifiers(skill, rules)
	found := false
	for _, tag := range skill.Tags {
		if tag == "security" {
			found = true
		}
	}
	if !found {
		t.Errorf("classifier tag 'security' should be appended: %v", skill.Tags)
	}
}

func TestApplyClassifiers_NoMatchDoesNothing(t *testing.T) {
	skill := &model.SkillRef{RepoPath: "docs/guide", Priority: 5}
	rules := []config.ClassifierConfig{
		{PathPrefix: "skills/security", Priority: 9},
	}
	applyClassifiers(skill, rules)
	if skill.Priority != 5 {
		t.Errorf("non-matching classifier should not change priority")
	}
}

func TestApplyClassifiers_EmptyPrefixSkipped(t *testing.T) {
	skill := &model.SkillRef{RepoPath: "skills/x", Priority: 5}
	rules := []config.ClassifierConfig{
		{PathPrefix: "", Priority: 99},
	}
	applyClassifiers(skill, rules)
	if skill.Priority != 5 {
		t.Errorf("empty prefix classifier should be skipped")
	}
}
