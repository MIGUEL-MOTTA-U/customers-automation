package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type GitHubRepo struct {
	Owner string
	Repo  string
}

type GitHubClient struct {
	httpClient *http.Client
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{Timeout: 40 * time.Second},
	}
}

func ParseGitHubRepo(repoURL string) (GitHubRepo, error) {
	u, err := url.Parse(strings.TrimSpace(repoURL))
	if err != nil {
		return GitHubRepo{}, fmt.Errorf("url inválida: %w", err)
	}
	if !strings.EqualFold(u.Host, "github.com") {
		return GitHubRepo{}, fmt.Errorf("host no soportado %q (solo github.com)", u.Host)
	}

	cleaned := strings.Trim(strings.TrimSpace(u.Path), "/")
	parts := strings.Split(cleaned, "/")
	if len(parts) < 2 {
		return GitHubRepo{}, fmt.Errorf("url de repositorio inválida: %q", repoURL)
	}
	return GitHubRepo{
		Owner: parts[0],
		Repo:  parts[1],
	}, nil
}

type repoInfoResponse struct {
	DefaultBranch string `json:"default_branch"`
}

func (c *GitHubClient) DefaultBranch(repo GitHubRepo) (string, error) {
	var info repoInfoResponse
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s", repo.Owner, repo.Repo)
	if err := c.getJSON(endpoint, &info); err != nil {
		return "", err
	}
	if info.DefaultBranch == "" {
		return "", fmt.Errorf("el repositorio %s/%s no reporta default_branch", repo.Owner, repo.Repo)
	}
	return info.DefaultBranch, nil
}

type treeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type treeResponse struct {
	Tree []treeEntry `json:"tree"`
}

func (c *GitHubClient) RepoTree(repo GitHubRepo, ref string) ([]treeEntry, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", repo.Owner, repo.Repo, url.PathEscape(ref))
	var tr treeResponse
	if err := c.getJSON(endpoint, &tr); err != nil {
		return nil, err
	}
	return tr.Tree, nil
}

type ContentEntry struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
}

func (c *GitHubClient) ListDir(repo GitHubRepo, ref, dirPath string) ([]ContentEntry, error) {
	endpoint := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		repo.Owner, repo.Repo, strings.TrimPrefix(path.Clean(dirPath), "/"), url.QueryEscape(ref),
	)
	var entries []ContentEntry
	if err := c.getJSON(endpoint, &entries); err == nil {
		return entries, nil
	}

	var single ContentEntry
	if err := c.getJSON(endpoint, &single); err != nil {
		return nil, err
	}
	return []ContentEntry{single}, nil
}

func (c *GitHubClient) DownloadRawFile(downloadURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request de descarga: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.raw")
	req.Header.Set("User-Agent", "skillsctl")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de red descargando %q: %w", downloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 400))
		return nil, fmt.Errorf("error HTTP %d descargando %q: %s", resp.StatusCode, downloadURL, string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo contenido descargado: %w", err)
	}
	return content, nil
}

func (c *GitHubClient) getJSON(endpoint string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("error creando request a github: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "skillsctl")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de red consultando github: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return fmt.Errorf("error HTTP %d en %q: %s", resp.StatusCode, endpoint, strings.TrimSpace(string(body)))
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("respuesta JSON inválida desde %q: %w", endpoint, err)
	}
	return nil
}
