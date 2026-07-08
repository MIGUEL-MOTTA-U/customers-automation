package install

import "testing"

func TestNormalizeLocalName_SpaceToHyphen(t *testing.T) {
	got := normalizeLocalName("My Skill")
	if got != "my-skill" {
		t.Errorf("got %q, want %q", got, "my-skill")
	}
}

func TestNormalizeLocalName_UnderscoreToHyphen(t *testing.T) {
	got := normalizeLocalName("my_skill")
	if got != "my-skill" {
		t.Errorf("got %q, want %q", got, "my-skill")
	}
}

func TestNormalizeLocalName_ForwardSlashToHyphen(t *testing.T) {
	got := normalizeLocalName("a/b/c")
	if got != "a-b-c" {
		t.Errorf("got %q, want %q", got, "a-b-c")
	}
}

func TestNormalizeLocalName_BackslashToHyphen(t *testing.T) {
	got := normalizeLocalName("a\\b")
	if got != "a-b" {
		t.Errorf("got %q, want %q", got, "a-b")
	}
}

func TestNormalizeLocalName_ConsecutiveHyphensCollapsed(t *testing.T) {
	got := normalizeLocalName("my--skill")
	if got != "my-skill" {
		t.Errorf("consecutive hyphens should collapse: got %q", got)
	}
}

func TestNormalizeLocalName_LeadingTrailingTrimmed(t *testing.T) {
	got := normalizeLocalName(" trim ")
	if got != "trim" {
		t.Errorf("leading/trailing spaces should be trimmed: got %q", got)
	}
}

func TestNormalizeLocalName_EmptyFallback(t *testing.T) {
	got := normalizeLocalName("")
	if got != "skill" {
		t.Errorf("empty input should return 'skill', got %q", got)
	}
}

func TestNormalizeLocalName_AllUppercase(t *testing.T) {
	got := normalizeLocalName("UPPERCASE")
	if got != "uppercase" {
		t.Errorf("should lowercase: got %q", got)
	}
}

func TestNormalizeLocalName_ComplexPath(t *testing.T) {
	got := normalizeLocalName("My / Complex_Skill")
	if got != "my-complex-skill" {
		t.Errorf("got %q, want %q", got, "my-complex-skill")
	}
}

func TestNormalizeLocalName_AlreadyNormalized(t *testing.T) {
	got := normalizeLocalName("simple-skill")
	if got != "simple-skill" {
		t.Errorf("already normalized should be unchanged: got %q", got)
	}
}
