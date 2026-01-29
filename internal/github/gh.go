package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Issue represents a GitHub issue
type Issue struct {
	Number int     `json:"number"`
	Title  string  `json:"title"`
	Body   string  `json:"body"`
	Labels []Label `json:"labels"`
	URL    string  `json:"url"`
}

// Label represents a GitHub label
type Label struct {
	Name string `json:"name"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number int           `json:"number"`
	Title  string        `json:"title"`
	Body   string        `json:"body"`
	Author Author        `json:"author"`
	State  string        `json:"state"`
	URL    string        `json:"url"`
	Files  []ChangedFile `json:"files"`
}

// Author represents a GitHub user
type Author struct {
	Login string `json:"login"`
}

// ChangedFile represents a file changed in a PR
type ChangedFile struct {
	Path string `json:"path"`
}

// GetIssue fetches an issue by number
func GetIssue(number int) (*Issue, error) {
	cmd := exec.Command("gh", "issue", "view", fmt.Sprintf("%d", number),
		"--json", "number,title,body,labels,url")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("failed to fetch issue #%d: %s", number, stderrStr)
		}
		return nil, fmt.Errorf("failed to fetch issue #%d: %w", number, err)
	}

	var issue Issue
	if err := json.Unmarshal(stdout.Bytes(), &issue); err != nil {
		return nil, fmt.Errorf("failed to parse issue response: %w", err)
	}

	return &issue, nil
}

// GetPullRequest fetches a PR by number
func GetPullRequest(number int) (*PullRequest, error) {
	cmd := exec.Command("gh", "pr", "view", fmt.Sprintf("%d", number),
		"--json", "number,title,body,author,state,url,files")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("failed to fetch PR #%d: %s", number, stderrStr)
		}
		return nil, fmt.Errorf("failed to fetch PR #%d: %w", number, err)
	}

	var pr PullRequest
	if err := json.Unmarshal(stdout.Bytes(), &pr); err != nil {
		return nil, fmt.Errorf("failed to parse PR response: %w", err)
	}

	return &pr, nil
}

// GHAvailable checks if gh CLI is installed and authenticated
func GHAvailable() bool {
	cmd := exec.Command("gh", "auth", "status")
	return cmd.Run() == nil
}

// Slugify converts a string to a URL-friendly slug
func Slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	s = reg.ReplaceAllString(s, "")

	// Replace multiple hyphens with single hyphen
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")

	// Trim hyphens from ends
	s = strings.Trim(s, "-")

	// Limit length
	if len(s) > 50 {
		s = s[:50]
		// Don't end with a hyphen
		s = strings.TrimSuffix(s, "-")
	}

	return s
}

// GenerateBranchName generates a branch name from issue/PR using hardcoded format
// Format: prefix-number-titleslug (e.g., "issue-42-fix-login-bug")
func GenerateBranchName(prefix string, number int, title string) string {
	slug := Slugify(title)
	return fmt.Sprintf("%s-%d-%s", prefix, number, slug)
}

// GetLabelNames returns a slice of label names
func (i *Issue) GetLabelNames() []string {
	names := make([]string, len(i.Labels))
	for idx, label := range i.Labels {
		names[idx] = label.Name
	}
	return names
}
