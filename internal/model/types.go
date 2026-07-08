package model

type SkillRef struct {
	ID        string   `json:"id" yaml:"id"`
	Name      string   `json:"name" yaml:"name"`
	SourceID  string   `json:"source_id" yaml:"source_id"`
	RepoURL   string   `json:"repo_url" yaml:"repo_url"`
	RepoPath  string   `json:"repo_path" yaml:"repo_path"`
	Ref       string   `json:"ref" yaml:"ref"`
	Tags      []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Roles     []string `json:"roles,omitempty" yaml:"roles,omitempty"`
	UseCases  []string `json:"use_cases,omitempty" yaml:"use_cases,omitempty"`
	Areas     []string `json:"areas,omitempty" yaml:"areas,omitempty"`
	Priority  int      `json:"priority,omitempty" yaml:"priority,omitempty"`
	Projects  []string `json:"projects,omitempty" yaml:"projects,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

type SelectionFile struct {
	Version   string     `json:"version" yaml:"version"`
	Project   string     `json:"project" yaml:"project"`
	CreatedAt string     `json:"created_at" yaml:"created_at"`
	Skills    []SkillRef `json:"skills" yaml:"skills"`
}

type InstallRecord struct {
	SkillID    string `json:"skill_id" yaml:"skill_id"`
	SourceID   string `json:"source_id" yaml:"source_id"`
	RepoURL    string `json:"repo_url" yaml:"repo_url"`
	RepoPath   string `json:"repo_path" yaml:"repo_path"`
	Ref        string `json:"ref" yaml:"ref"`
	LocalPath  string `json:"local_path" yaml:"local_path"`
	Installed  string `json:"installed_at" yaml:"installed_at"`
	LastUpdate string `json:"last_update_at" yaml:"last_update_at"`
}

type InstallState struct {
	Version string          `json:"version" yaml:"version"`
	Project string          `json:"project" yaml:"project"`
	Records []InstallRecord `json:"records" yaml:"records"`
}
