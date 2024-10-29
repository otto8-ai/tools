package documentloader

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"log/slog"
	"os"
	"strings"
)

// CloneRepo clones a git repository to a target directory
// repo is the repository to clone - may contain an @ symbol to specify a commit, tag or branch (prioritized in that order)
func CloneRepo(repo, target string) error {
	atSplit := strings.Split(repo, "@")

	if len(atSplit) > 2 {
		return fmt.Errorf("invalid repository URL format %q", repo)
	}

	slog.Info("Cloning repository", "repo", repo)

	opts := &git.CloneOptions{
		URL:      atSplit[0],
		Progress: os.Stderr,
	}

	// Clone the repository to the temporary directory
	r, err := git.PlainClone(target, false, opts)
	if err != nil {
		return fmt.Errorf("failed to clone repo %q: %w", repo, err)
	}

	if len(atSplit) == 1 {
		return nil
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	checkoutOpts := &git.CheckoutOptions{
		Create: false,
		Force:  true,
	}

	// Check out the specific branch, tag, or commit
	if len(atSplit[1]) == 40 && isHex(atSplit[1]) {
		checkoutOpts.Hash = plumbing.NewHash(atSplit[1])
		slog.Info("Checking out commit", "commit", atSplit[1])
	} else {
		checkoutOpts.Branch = plumbing.NewTagReferenceName(atSplit[1])
		if err := w.Checkout(checkoutOpts); err == nil {
			slog.Info("Checked out tag", "tag", atSplit[1])
			return nil
		}
		// failed as tag - try as branch
		checkoutOpts.Branch = plumbing.NewBranchReferenceName(atSplit[1])
		slog.Info("Checking out branch", "branch", atSplit[1])
	}

	if err := w.Checkout(checkoutOpts); err != nil {
		return fmt.Errorf("failed to checkout %q: %w", atSplit[1], err)
	}

	return nil
}

func isHex(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
