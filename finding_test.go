package main

import (
	"path/filepath"
	"testing"
)

func TestFindingAddListLink(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "findings")

	if err := findingAdd(dir, []string{"SQL", "Injection", "in", "login", "--severity", "high"}); err != nil {
		t.Fatal(err)
	}
	if err := findingAdd(dir, []string{"Reflected", "XSS"}); err != nil {
		t.Fatal(err)
	}

	names, err := findingFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("got %d finding files, want 2", len(names))
	}
	if names[0][:4] != "0001" || names[1][:4] != "0002" {
		t.Fatalf("unexpected ids: %v", names)
	}

	f, err := parseFinding(filepath.Join(dir, names[0]))
	if err != nil {
		t.Fatal(err)
	}
	if f.Title != "SQL Injection in login" || f.Severity != "high" {
		t.Fatalf("unexpected parse: %+v", f)
	}

	if err := findingLink(dir, []string{"1", "evidence/exploitation/shot.png"}); err != nil {
		t.Fatal(err)
	}
	f, err = parseFinding(filepath.Join(dir, names[0]))
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Evidence) != 1 || f.Evidence[0] != "evidence/exploitation/shot.png" {
		t.Fatalf("evidence not linked: %+v", f)
	}
	if f.Body == "" {
		t.Fatal("expected body content to survive the link insert")
	}
}
