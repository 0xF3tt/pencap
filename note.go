package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func cmdNote(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: pencap note <type> <text...>")
	}
	typ := args[0]
	if err := validateName("note type", typ); err != nil {
		return err
	}
	text := strings.Join(args[1:], " ")

	root, err := findRoot()
	if err != nil {
		return err
	}

	dir := filepath.Join(root, "notes")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	path := filepath.Join(dir, typ+".md")
	line := fmt.Sprintf("- %s %s\n", time.Now().UTC().Format(time.RFC3339), text)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304 G703
	if err != nil {
		return err
	}

	if _, err := f.WriteString(line); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	fmt.Println("saved", path)
	return nil
}
