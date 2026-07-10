package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRoot(t *testing.T) {
	tmp := t.TempDir()
	eng := filepath.Join(tmp, "acme-2026")
	sub := filepath.Join(eng, "evidence", "recon")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(eng, markerFile), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()

	if err := os.Chdir(sub); err != nil {
		t.Fatal(err)
	}
	root, err := findRoot()
	if err != nil {
		t.Fatal(err)
	}
	wantRoot, err := filepath.EvalSymlinks(eng)
	if err != nil {
		t.Fatal(err)
	}
	if root != wantRoot {
		t.Fatalf("got %s, want %s", root, wantRoot)
	}
}

func TestFindRootMissing(t *testing.T) {
	tmp := t.TempDir()
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	if _, err := findRoot(); err == nil {
		t.Fatal("expected error when no scope.yaml present")
	}
}

func TestValidateName(t *testing.T) {
	bad := []string{"", "../etc", "a/b", "..", "foo/../bar"}
	for _, s := range bad {
		if err := validateName("test", s); err == nil {
			t.Errorf("validateName(%q) = nil, want error", s)
		}
	}

	good := []string{"acme-2026", "recon", "file", "notes.v2"}
	for _, s := range good {
		if err := validateName("test", s); err != nil {
			t.Errorf("validateName(%q) = %v, want nil", s, err)
		}
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"":                   "shot",
		"admin panel login!": "admin-panel-login",
		"  spaced  ":         "spaced",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
