//go:build !windows

package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// addToPath appends an export line to the user's shell rc file so dir is on
// PATH in future shells. Idempotent: does nothing if already added.
func addToPath(dir string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	rcCandidates := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".profile"),
	}

	target := ""
	for _, rc := range rcCandidates {
		if _, err := os.Stat(rc); err == nil {
			target = rc
			break
		}
	}
	if target == "" {
		target = filepath.Join(home, ".profile")
	}

	marker := fmt.Sprintf("# added by diagram-mcp installer (%s)", dir)
	if existing, err := os.ReadFile(target); err == nil {
		if strings.Contains(string(existing), marker) {
			return nil // already added
		}
	}

	f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("\n%s\nexport PATH=\"%s:$PATH\"\n", marker, dir)
	_, err = f.WriteString(line)
	return err
}
