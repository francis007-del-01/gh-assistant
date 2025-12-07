package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Git provides git operations
type Git struct {
	workDir string
}

// New creates a new Git instance
func New(workDir string) *Git {
	if workDir == "" {
		workDir = "."
	}
	return &Git{workDir: workDir}
}

// run executes a git command and returns the output
func (g *Git) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// IsRepo checks if the current directory is a git repository
func (g *Git) IsRepo() bool {
	_, err := g.run("rev-parse", "--git-dir")
	return err == nil
}

// GetStagedDiff returns the diff of staged changes
func (g *Git) GetStagedDiff() (string, error) {
	return g.run("diff", "--cached")
}

// GetUnstagedDiff returns the diff of unstaged changes
func (g *Git) GetUnstagedDiff() (string, error) {
	return g.run("diff")
}

// GetAllDiff returns all changes (staged + unstaged)
func (g *Git) GetAllDiff() (string, error) {
	return g.run("diff", "HEAD")
}

// GetUnpushedCommits returns commits that haven't been pushed
func (g *Git) GetUnpushedCommits() ([]string, error) {
	branch, err := g.GetCurrentBranch()
	if err != nil {
		return nil, err
	}

	// Get the upstream branch
	upstream, err := g.run("rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if err != nil {
		// No upstream set, return empty
		return nil, nil
	}

	// Get unpushed commit hashes
	output, err := g.run("log", upstream+"..HEAD", "--format=%H")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

// GetUnpushedCommitMessages returns commit messages for unpushed commits
// Format: ["hash:message", "hash:message", ...]
func (g *Git) GetUnpushedCommitMessages() ([]string, error) {
	branch, err := g.GetCurrentBranch()
	if err != nil {
		return nil, err
	}

	// Get the upstream branch
	upstream, err := g.run("rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if err != nil {
		// No upstream set, return empty
		return nil, nil
	}

	// Get unpushed commits with short hash and subject
	output, err := g.run("log", upstream+"..HEAD", "--format=%h - %s")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

// GetCommitDiff returns the diff for a specific commit
func (g *Git) GetCommitDiff(commitHash string) (string, error) {
	return g.run("show", commitHash, "--format=", "--no-color")
}

// GetUnpushedDiff returns combined diff of all unpushed commits
func (g *Git) GetUnpushedDiff() (string, error) {
	branch, err := g.GetCurrentBranch()
	if err != nil {
		return "", err
	}

	upstream, err := g.run("rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if err != nil {
		// No upstream, get diff against empty tree (all changes)
		return g.run("diff", "4b825dc642cb6eb9a060e54bf8d69288fbee4904..HEAD")
	}

	return g.run("diff", upstream+"..HEAD")
}

// GetCurrentBranch returns the current branch name
func (g *Git) GetCurrentBranch() (string, error) {
	return g.run("rev-parse", "--abbrev-ref", "HEAD")
}

// GetRemote returns the default remote (usually "origin")
func (g *Git) GetRemote() (string, error) {
	output, err := g.run("remote")
	if err != nil {
		return "", err
	}

	remotes := strings.Split(output, "\n")
	if len(remotes) == 0 {
		return "", errors.New("no remote configured")
	}

	// Prefer "origin" if available
	for _, r := range remotes {
		if r == "origin" {
			return r, nil
		}
	}

	return remotes[0], nil
}

// HasStagedChanges checks if there are staged changes
func (g *Git) HasStagedChanges() (bool, error) {
	output, err := g.run("diff", "--cached", "--name-only")
	if err != nil {
		return false, err
	}
	return output != "", nil
}

// HasUnstagedChanges checks if there are unstaged changes
func (g *Git) HasUnstagedChanges() (bool, error) {
	output, err := g.run("diff", "--name-only")
	if err != nil {
		return false, err
	}
	return output != "", nil
}

// StageAll stages all changes
func (g *Git) StageAll() error {
	_, err := g.run("add", "-A")
	return err
}

// Commit creates a commit with the given message
func (g *Git) Commit(message string) error {
	_, err := g.run("commit", "-m", message)
	return err
}

// AmendCommit amends the last commit with a new message
func (g *Git) AmendCommit(message string) error {
	_, err := g.run("commit", "--amend", "-m", message)
	return err
}

// Push pushes to the remote
func (g *Git) Push() error {
	remote, err := g.GetRemote()
	if err != nil {
		return err
	}

	branch, err := g.GetCurrentBranch()
	if err != nil {
		return err
	}

	_, err = g.run("push", remote, branch)
	return err
}

// PushSetUpstream pushes and sets upstream
func (g *Git) PushSetUpstream() error {
	remote, err := g.GetRemote()
	if err != nil {
		return err
	}

	branch, err := g.GetCurrentBranch()
	if err != nil {
		return err
	}

	_, err = g.run("push", "-u", remote, branch)
	return err
}

// GetStatus returns the git status
func (g *Git) GetStatus() (string, error) {
	return g.run("status", "--short")
}

// GetLastCommitMessage returns the message of the last commit
func (g *Git) GetLastCommitMessage() (string, error) {
	return g.run("log", "-1", "--format=%B")
}

// GetChangedFiles returns a list of changed files
func (g *Git) GetChangedFiles() ([]string, error) {
	output, err := g.run("diff", "--name-only", "HEAD")
	if err != nil {
		// Try without HEAD for initial commit
		output, err = g.run("diff", "--cached", "--name-only")
		if err != nil {
			return nil, err
		}
	}

	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

// IsFirstPushToBranch checks if the current branch has no upstream tracking branch
// This indicates it's a new branch that hasn't been pushed yet
func (g *Git) IsFirstPushToBranch() (bool, error) {
	branch, err := g.GetCurrentBranch()
	if err != nil {
		return false, err
	}

	// Try to get the upstream branch
	_, err = g.run("rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if err != nil {
		// No upstream means this is a first push
		return true, nil
	}

	return false, nil
}

// IsMainBranch checks if the current branch is main or master
func (g *Git) IsMainBranch() bool {
	branch, err := g.GetCurrentBranch()
	if err != nil {
		return false
	}
	return branch == "main" || branch == "master"
}

