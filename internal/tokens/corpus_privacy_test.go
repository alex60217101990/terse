package tokens_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The corpus is committed to a public repository. A sample harvested from a
// real transcript archive would leak home paths, private source, and possibly
// secrets. Discipline is not enough — check it.
func TestCorpusHasNoRealUserData(t *testing.T) {
	banned := []string{
		"/Users/", "/home/", "C:\\Users\\", "ssh-rsa", "BEGIN PRIVATE KEY",
		"xoxb-", "ghp_", "sk-", "AKIA",
	}
	root := corpusRoot
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		for _, bad := range banned {
			if strings.Contains(string(b), bad) {
				t.Errorf("%s contains %q — corpus samples must be invented, "+
					"never harvested from a real archive", p, bad)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
}
