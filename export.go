package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func cmdExport(args []string) error {
	root, err := findRoot()
	if err != nil {
		return err
	}

	names, err := findingFiles(filepath.Join(root, "findings"))
	if err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Pentest Report\n\ngenerated: %s\n\n", time.Now().UTC().Format(time.RFC3339))

	if len(names) == 0 {
		b.WriteString("_no findings recorded yet_\n")
	}

	for _, n := range names {
		f, err := parseFinding(filepath.Join(root, "findings", n))
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "## [%s] %s (%s)\n\n", f.ID, f.Title, strings.ToUpper(f.Severity))
		if f.Body != "" {
			fmt.Fprintf(&b, "%s\n\n", f.Body)
		}
		if len(f.Evidence) > 0 {
			b.WriteString("Evidence:\n\n")
			for _, e := range f.Evidence {
				fmt.Fprintf(&b, "- `%s`\n", e)
			}
			b.WriteString("\n")
		}
	}

	dir := filepath.Join(root, "report", "draft")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	path := filepath.Join(dir, "report.md")
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		return err
	}

	fmt.Println("saved", path)
	return nil
}
