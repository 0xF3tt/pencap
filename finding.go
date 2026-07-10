package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var findingFileRe = regexp.MustCompile(`^(\d{4})-`)

var validSeverities = map[string]bool{"crit": true, "high": true, "med": true, "low": true, "info": true}

type finding struct {
	ID       string
	Title    string
	Severity string
	Created  string
	Evidence []string
	Body     string
}

func cmdFinding(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pencap finding add <title> [--severity <crit|high|med|low|info>] | pencap finding link <id> <evidence-path> | pencap finding list")
	}

	root, err := findRoot()
	if err != nil {
		return err
	}
	dir := filepath.Join(root, "findings")

	switch args[0] {
	case "add":
		return findingAdd(dir, args[1:])
	case "link":
		return findingLink(dir, args[1:])
	case "list":
		return findingList(dir)
	default:
		return fmt.Errorf("unknown finding subcommand %q", args[0])
	}
}

func findingAdd(dir string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pencap finding add <title> [--severity <crit|high|med|low|info>]")
	}

	severity := "info"
	var titleParts []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--severity" && i+1 < len(args) {
			severity = args[i+1]
			i++
			continue
		}
		titleParts = append(titleParts, args[i])
	}
	title := strings.Join(titleParts, " ")
	if title == "" {
		return fmt.Errorf("title required")
	}
	if !validSeverities[severity] {
		return fmt.Errorf("invalid severity %q: must be one of crit, high, med, low, info", severity)
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	id, err := nextFindingID(dir)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, fmt.Sprintf("%04d-%s.md", id, slugify(title)))
	content := fmt.Sprintf("id: %04d\ntitle: %s\nseverity: %s\ncreated: %s\n\n## Description\n\n",
		id, title, severity, time.Now().UTC().Format(time.RFC3339))

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil { // #nosec G703
		return err
	}

	fmt.Println("saved", path)
	return nil
}

func nextFindingID(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	max := 0
	for _, e := range entries {
		m := findingFileRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		if n, err := strconv.Atoi(m[1]); err == nil && n > max {
			max = n
		}
	}
	return max + 1, nil
}

func findingPath(dir, id string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	padded := id
	if n, err := strconv.Atoi(id); err == nil {
		padded = fmt.Sprintf("%04d", n)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), padded+"-") {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no finding with id %s", id)
}

func findingLink(dir string, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: pencap finding link <id> <evidence-path>")
	}
	path, err := findingPath(dir, args[0])
	if err != nil {
		return err
	}

	// path came from findingPath(), which only returns names already listed
	// by os.ReadDir — never raw, unvalidated input.
	b, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")

	// evidence lines are inserted right before the header/body blank-line
	// separator so `finding add`'s output stays a valid insertion point.
	insertAt := len(lines)
	for i, l := range lines {
		if l == "" {
			insertAt = i
			break
		}
	}
	line := "evidence: " + args[1]
	lines = append(lines[:insertAt], append([]string{line}, lines[insertAt:]...)...)

	// path came from findingPath(), which only returns names already listed
	// by os.ReadDir — never raw, unvalidated input.
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600); err != nil { // #nosec G703
		return err
	}
	fmt.Println("linked", args[1], "to", filepath.Base(path))
	return nil
}

func parseFinding(path string) (finding, error) {
	// path is always one of findingFiles()'s os.ReadDir results — never
	// raw, unvalidated input.
	b, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return finding{}, err
	}
	lines := strings.Split(string(b), "\n")

	var f finding
	i := 0
	for ; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			i++
			break
		}
		switch {
		case strings.HasPrefix(line, "id: "):
			f.ID = strings.TrimPrefix(line, "id: ")
		case strings.HasPrefix(line, "title: "):
			f.Title = strings.TrimPrefix(line, "title: ")
		case strings.HasPrefix(line, "severity: "):
			f.Severity = strings.TrimPrefix(line, "severity: ")
		case strings.HasPrefix(line, "created: "):
			f.Created = strings.TrimPrefix(line, "created: ")
		case strings.HasPrefix(line, "evidence: "):
			f.Evidence = append(f.Evidence, strings.TrimPrefix(line, "evidence: "))
		}
	}
	f.Body = strings.TrimRight(strings.Join(lines[i:], "\n"), "\n")
	return f, nil
}

func findingFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func findingList(dir string) error {
	names, err := findingFiles(dir)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Println("no findings yet")
		return nil
	}
	for _, n := range names {
		f, err := parseFinding(filepath.Join(dir, n))
		if err != nil {
			return err
		}
		fmt.Printf("%-4s  %-8s  %-40s  (%d evidence)\n", f.ID, f.Severity, f.Title, len(f.Evidence))
	}
	return nil
}
