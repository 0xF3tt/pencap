package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const markerFile = "scope.yaml"

var evidenceTypes = []string{"recon", "staging", "exploitation", "postex", "files"}

func validateName(kind, s string) error {
	if s == "" || s != filepath.Base(s) || strings.Contains(s, "..") {
		return fmt.Errorf("invalid %s %q: must be a plain name with no path separators", kind, s)
	}
	return nil
}

func cmdInit(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: pencap init <engagement-name>")
	}
	name := args[0]
	if err := validateName("engagement name", name); err != nil {
		return err
	}

	if _, err := os.Stat(name); err == nil { // #nosec G703
		return fmt.Errorf("%s already exists", name)
	}

	dirs := []string{"notes", "findings", "report/draft", "report/final"}
	for _, t := range evidenceTypes {
		dirs = append(dirs, filepath.Join("evidence", t))
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(name, d), 0o750); err != nil { // #nosec G703
			return err
		}
	}

	scope := `# scope.yaml - engagement scope and rules of engagement
engagement: ` + name + `
client: ""
start_date: ""
end_date: ""
in_scope:
  - ""
out_of_scope:
  - ""
contacts:
  - ""
`
	if err := os.WriteFile(filepath.Join(name, markerFile), []byte(scope), 0o600); err != nil { // #nosec G703
		return err
	}

	fmt.Printf("engagement scaffolded at %s\n", name)
	return nil
}

// findRoot resolves the active engagement: the nearest enclosing directory with
// a scope.yaml wins (local context first), falling back to the globally
// configured path (`pencap path`) when you're outside any engagement tree.
func findRoot() (string, error) {
	if dir, err := findRootFromCwd(); err == nil {
		return dir, nil
	}
	if cfg, err := loadConfigRoot(); err == nil && cfg != "" {
		if _, err := os.Stat(filepath.Join(cfg, markerFile)); err == nil {
			return cfg, nil
		}
	}
	return "", fmt.Errorf("not inside an engagement (no %s found) and no valid global path set; run `pencap init`, then `pencap path <dir>`", markerFile)
}

func findRootFromCwd() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, markerFile)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside an engagement (no %s found)", markerFile)
		}
		dir = parent
	}
}
