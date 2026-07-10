package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type sidecar struct {
	Timestamp string `json:"timestamp"`
	Host      string `json:"host"`
	Type      string `json:"type"`
	Note      string `json:"note,omitempty"`
	Source    string `json:"source,omitempty"`
	SHA256    string `json:"sha256"`
}

// fileSHA256 gives evidence files a checksum for chain-of-custody: proof the
// file handed to the client's report hasn't changed since capture.
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path) // #nosec G304 G703
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

var slugRe = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func slugify(s string) string {
	s = slugRe.ReplaceAllString(strings.TrimSpace(s), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "shot"
	}
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}

func cmdScreenshot(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pencap ss <type> [note...] | pencap ss file <src-path> [note...]")
	}
	typ := args[0]
	rest := args[1:]
	if err := validateName("evidence type", typ); err != nil {
		return err
	}

	root, err := findRoot()
	if err != nil {
		return err
	}

	if typ == "file" {
		if len(rest) < 1 {
			return fmt.Errorf("usage: pencap ss file <src-path> [note...]")
		}
		return copyFileEvidence(root, rest[0], strings.Join(rest[1:], " "))
	}

	return takeScreenshot(root, typ, strings.Join(rest, " "))
}

func writeSidecar(destPath string, sc sidecar) error {
	b, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(destPath+".json", b, 0o600) // #nosec G703
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

func takeScreenshot(root, typ, note string) error {
	cmd, cmdArgs, err := screenshotCommand()
	if err != nil {
		return err
	}

	dir := filepath.Join(root, "evidence", typ)
	if err := os.MkdirAll(dir, 0o750); err != nil { // #nosec G703
		return err
	}

	ts := time.Now().UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("%s_%s.png", ts, slugify(note))
	dest := filepath.Join(dir, name)

	fullArgs := append(append([]string{}, cmdArgs...), dest)
	c := exec.Command(cmd, fullArgs...) // #nosec G204 G702
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("screenshot capture failed: %w", err)
	}

	if _, err := os.Stat(dest); err != nil { // #nosec G703 -- dest built from validated typ above
		return fmt.Errorf("no screenshot saved (selection cancelled?)")
	}

	sum, err := fileSHA256(dest)
	if err != nil {
		return err
	}
	if err := writeSidecar(dest, sidecar{Timestamp: ts, Host: hostname(), Type: typ, Note: note, SHA256: sum}); err != nil {
		return err
	}

	fmt.Println("saved", dest)
	return nil
}

func screenshotCommand() (string, []string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "screencapture", []string{"-i"}, nil
	case "linux":
		opts := []struct {
			bin  string
			args []string
		}{
			{"scrot", []string{"-s"}},
			{"gnome-screenshot", []string{"-a", "-f"}},
			{"import", nil},
		}
		for _, o := range opts {
			if _, err := exec.LookPath(o.bin); err == nil {
				return o.bin, o.args, nil
			}
		}
		return "", nil, fmt.Errorf("no screenshot tool found (install scrot, gnome-screenshot, or imagemagick)")
	default:
		return "", nil, fmt.Errorf("screenshot capture not supported on %s; use `pencap ss file <path>` to import one instead", runtime.GOOS)
	}
}

func copyFileEvidence(root, src, note string) error {
	in, err := os.Open(src) // #nosec G304 G703
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	dir := filepath.Join(root, "evidence", "files")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	ts := time.Now().UTC().Format("20060102T150405Z")
	dest := filepath.Join(dir, fmt.Sprintf("%s_%s", ts, filepath.Base(src)))

	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	sum, err := fileSHA256(dest)
	if err != nil {
		return err
	}
	if err := writeSidecar(dest, sidecar{Timestamp: ts, Host: hostname(), Type: "file", Note: note, Source: src, SHA256: sum}); err != nil {
		return err
	}

	fmt.Println("saved", dest)
	return nil
}
