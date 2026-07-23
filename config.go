package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// configPath is the file holding the globally-configured engagement root.
// PENCAP_CONFIG overrides it (used by tests and to point at an alternate file).
func configPath() (string, error) {
	if p := os.Getenv("PENCAP_CONFIG"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "pencap", "config"), nil
}

// loadConfigRoot returns the globally-configured engagement root, or "" if unset.
func loadConfigRoot() (string, error) {
	p, err := configPath()
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(p) // #nosec G304 -- path from PENCAP_CONFIG or os.UserConfigDir, not attacker-controlled
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func saveConfigRoot(dir string) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil { // #nosec G703
		return err
	}
	return os.WriteFile(p, []byte(dir+"\n"), 0o600) // #nosec G703
}

// engagementName reads the `engagement:` field from an engagement's scope.yaml.
func engagementName(root string) string {
	b, err := os.ReadFile(filepath.Join(root, markerFile)) // #nosec G304 -- root resolved by findRoot
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(b), "\n") {
		if v, ok := strings.CutPrefix(line, "engagement:"); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func cmdPath(args []string) error {
	if len(args) == 0 {
		root, err := loadConfigRoot()
		if err != nil {
			return err
		}
		if root == "" {
			fmt.Println("no global engagement path set")
		} else {
			fmt.Println(root)
		}
		return nil
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: pencap path [<dir>]")
	}

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, markerFile)); err != nil {
		return fmt.Errorf("%s is not an engagement (no %s found)", dir, markerFile)
	}
	if err := saveConfigRoot(dir); err != nil {
		return err
	}
	fmt.Println("global engagement path set to", dir)
	return nil
}

func cmdInfo() error {
	cfg, err := loadConfigRoot()
	if err != nil {
		return err
	}
	if cfg == "" {
		fmt.Println("global path:       (unset)")
	} else {
		fmt.Printf("global path:       %s\n", cfg)
	}

	root, err := findRoot()
	if err != nil {
		fmt.Println("active engagement: none —", err)
		return nil
	}
	source := "global config"
	if local, err := findRootFromCwd(); err == nil && local == root {
		source = "current directory"
	}
	fmt.Printf("active engagement: %s\n", root)
	if name := engagementName(root); name != "" {
		fmt.Printf("name:              %s\n", name)
	}
	fmt.Printf("resolved via:      %s\n", source)
	return nil
}
