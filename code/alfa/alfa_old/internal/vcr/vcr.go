package vcr

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const MAIN = "A"

type Vcr struct {
	workdir string
	repo    *git.Repository
}

// NewVcr initializes or opens a repository
func NewVcr(user, workdir string) *Vcr {
	v := &Vcr{workdir: workdir}
	if user != "default" {
		v.initGit(user)
	}
	return v
}

func (v *Vcr) initGit(user string) {
	var err error
	v.repo, err = git.PlainOpen(v.workdir)
	if err == git.ErrRepositoryNotExists {
		v.repo, err = git.PlainInit(v.workdir, false)
		if err != nil {
			log.Fatalf("Failed to init repo: %v", err)
		}
		w, _ := v.repo.Worktree()
		v.repo.CreateBranch(&config.Branch{Name: MAIN})
		w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(MAIN), Create: true})
		v.Commit("Initialized")
	}
}

// Commit creates a commit with the given message and returns the commit id.
func (v *Vcr) Commit(message string) string {
	if v.repo == nil {
		return "?"
	}
	w, err := v.repo.Worktree()
	if err != nil {
		log.Printf("Cannot get worktree: %v", err)
		return ""
	}

	err = w.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		log.Printf("Add failed: %v", err)
	}

	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "VCR Bot",
			Email: "bot@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Printf("Commit failed: %v", err)
		return ""
	}

	// Auto-push to workbench remote after commit
	if err := v.Push("workbench"); err != nil {
		log.Printf("Warning: Push to workbench failed: %v", err)
	}

	return commit.String()
}

// Push pushes the current branch to the specified remote
func (v *Vcr) Push(remoteName string) error {
	if v.repo == nil {
		return fmt.Errorf("no repo initialized")
	}

	// Get current branch
	ref, err := v.repo.Head()
	if err != nil {
		return fmt.Errorf("cannot get HEAD: %w", err)
	}

	// Push to remote
	err = v.repo.Push(&git.PushOptions{
		RemoteName: remoteName,
		RefSpecs: []config.RefSpec{
			config.RefSpec(ref.Name() + ":" + ref.Name()),
		},
	})

	// Ignore "already up-to-date" errors
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("push failed: %w", err)
	}

	return nil
}

// branchFrom creates a new branch from a tag (or HEAD)
// BranchFrom creates a new branch from a tag or HEAD, with a comment.
func (v *Vcr) BranchFrom(baseTag, comment string) string {
	if v.repo == nil {
		return "?"
	}

	w, err := v.repo.Worktree()
	if err != nil {
		log.Printf("Cannot get worktree: %v", err)
		return ""
	}

	newBranch := v.nextBranchName()
	if newBranch == "" {
		return ""
	}

	ref := plumbing.NewBranchReferenceName(newBranch)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: ref,
		Create: true,
		Keep:   false,
	})
	if err != nil {
		log.Printf("Checkout failed: %v", err)
		return ""
	}

	tag := v.Commit(fmt.Sprintf("Branched from %s: %s", baseTag, comment))
	return tag
}

// nextBranchName returns next branch name like B, C, ...
func (v *Vcr) nextBranchName() string {
	iter, err := v.repo.Branches()
	if err != nil {
		log.Printf("Cannot list branches: %v", err)
		return ""
	}
	defer iter.Close()

	count := 0
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		if !strings.HasSuffix(ref.Name().String(), MAIN) {
			count++
		}
		return nil
	})
	if err != nil {
		log.Printf("Branch iteration error: %v", err)
	}
	if count > 25 {
		return ""
	}
	return string(rune('B' + count))
}

// getHistory returns a list of commits sorted by time
// GetHistory returns a list of commits sorted by time.
func (v *Vcr) GetHistory() []string {
	var result []string
	if v.repo == nil {
		return result
	}
	ref, err := v.repo.Head()
	if err != nil {
		log.Printf("Cannot get HEAD: %v", err)
		return result
	}
	commitIter, err := v.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		log.Printf("Cannot get commit log: %v", err)
		return result
	}
	defer commitIter.Close()

	err = commitIter.ForEach(func(c *object.Commit) error {
		result = append(result, fmt.Sprintf("%s|%s", c.Committer.When.Format("2006-01-02 15:04"), c.Message))
		return nil
	})
	if err != nil {
		log.Printf("Error iterating commits: %v", err)
	}
	sort.Strings(result)
	return result
}

// checkout switches to a given branch or tag
// Checkout switches to a given branch or tag.
func (v *Vcr) Checkout(refName string) error {
	if v.repo == nil {
		return fmt.Errorf("no repo initialized")
	}
	w, err := v.repo.Worktree()
	if err != nil {
		return err
	}
	return w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + refName),
	})
}

// rewriteToMain replaces main branch with content from ref
// RewriteToMain replaces main branch with content from ref and commits with a message.
func (v *Vcr) RewriteToMain(sourceRef string, message string) error {
	if v.repo == nil {
		return fmt.Errorf("no repo initialized")
	}
	w, err := v.repo.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(MAIN),
		Force:  true,
	})
	if err != nil {
		return err
	}

	hash, err := v.repo.ResolveRevision(plumbing.Revision(sourceRef))
	if err != nil {
		return err
	}
	err = w.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: *hash,
	})
	if err != nil {
		return err
	}
	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "VCR Bot",
			Email: "bot@example.com",
			When:  time.Now(),
		},
	})
	return err
}

// purge deletes the .git directory to reset versioning
// Purge deletes the .git directory to reset versioning
func (v *Vcr) Purge() error {
	gitDir := filepath.Join(v.workdir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		err := os.RemoveAll(gitDir)
		if err != nil {
			return fmt.Errorf("failed to remove .git: %w", err)
		}
	}
	return nil
}
