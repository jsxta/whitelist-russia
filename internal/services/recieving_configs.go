package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitService struct {
	LocalPath string
	RemoteURL string
	Branch    string
	APIUrl    string
}

func NewGitService(localPath, remoteURL, branch string) *GitService {
	api := toAPIUrl(remoteURL, branch)
	return &GitService{
		LocalPath: localPath,
		RemoteURL: remoteURL,
		Branch:    branch,
		APIUrl:    api,
	}
}
func (s GitService) Pull() error {
	if _, err := os.Stat(s.LocalPath); os.IsNotExist(err) {
		return s.cloneRepo()
	}

	r, err := git.PlainOpen(s.LocalPath)
	if err != nil {
		_ = os.RemoveAll(s.LocalPath)
		return s.cloneRepo()
	}

	if err := r.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec("+refs/heads/*:refs/remotes/origin/*")},
		Force:      true,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		_ = os.RemoveAll(s.LocalPath)
		return s.cloneRepo()
	}

	remoteRef := plumbing.NewRemoteReferenceName("origin", s.Branch)
	ref, err := r.Reference(remoteRef, true)
	if err != nil {
		_ = os.RemoveAll(s.LocalPath)
		return s.cloneRepo()
	}

	w, err := r.Worktree()
	if err != nil {
		_ = os.RemoveAll(s.LocalPath)
		return s.cloneRepo()
	}

	if err := w.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: ref.Hash(),
	}); err != nil {
		_ = os.RemoveAll(s.LocalPath)
		return s.cloneRepo()
	}

	return nil
}

func (s GitService) cloneRepo() error {
	if err := os.MkdirAll(filepath.Dir(s.LocalPath), 0o755); err != nil {
		return err
	}

	_, err := git.PlainClone(s.LocalPath, false, &git.CloneOptions{
		URL:           s.RemoteURL,
		ReferenceName: plumbing.NewBranchReferenceName(s.Branch),
		SingleBranch:  true,
		Depth:         1,
	})
	return err
}

func (s GitService) GetCurrentLocalCommit() (string, error) {
	r, err := git.PlainOpen(s.LocalPath)
	if err != nil {
		return "", err
	}
	hash, err := r.ResolveRevision(plumbing.Revision(s.Branch))
	if err != nil {
		return "", err
	}
	return hash.String(), nil
}

func (s GitService) GetLatestRemoteCommit() (string, error) {
	resp, err := http.Get(s.APIUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	data := make(map[string]any)
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	return data["sha"].(string), nil

}

func (s GitService) UpdateRepo() error {
	localCommit, err := s.GetCurrentLocalCommit()
	if err != nil {
		log.Printf("Error in local repo: %s: %s\n", s.LocalPath, err)
		s.Pull()
		return err
	}
	remoteCommit, err := s.GetLatestRemoteCommit()
	if err != nil {
		log.Printf("Error by getting remote commit: %s: %s\n", s.APIUrl, err)
		s.Pull()
		return err
	}
	if localCommit == remoteCommit {
		return nil
	}
	return s.Pull()

}

func (s GitService) ExistNewCommit() (bool, error) {
	localCommit, err := s.GetCurrentLocalCommit()
	if err != nil {
		return false, err
	}
	remoteCommit, err := s.GetLatestRemoteCommit()
	if err != nil {
		return false, err
	}
	if localCommit == remoteCommit {
		return false, nil
	}
	return true, nil
}

func toAPIUrl(raw, branch string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	owner := parts[0]
	repo := parts[1]
	repo = strings.TrimSuffix(repo, ".git")
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", owner, repo, branch)
	return api
}
