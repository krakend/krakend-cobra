package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/krakendio/krakend-cobra/v2/plugin"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

// indirectRequires returns the indirect dependencies of the go.sum file.
func indirectRequires(goSum string) (map[string]struct{}, error) {
	dir := filepath.Dir(goSum)
	filename := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read go.mod: %w", err)
	}

	f, err := modfile.Parse(filename, data, nil)
	if err != nil {
		return nil, fmt.Errorf("parse go.mod: %w", err)
	}

	indirects := map[string]struct{}{}
	for _, r := range f.Require {
		if r.Indirect {
			indirects[r.Mod.Path] = struct{}{}
		}
	}

	return indirects, nil
}

// getBuildInfo returns the dependencies of the binary calling it.
// It is a var to allow the replacement of the function in the tests
// as the debug.ReadBuildInfo function is not available in the tests
// https://github.com/golang/go/issues/68045
var localDescriber = plugin.Local

func pluginFunc(cmd *cobra.Command, _ []string) error {
	f, err := os.Open(goSum)
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }() // Workaround false positive for GO-S2307.

	desc, err := plugin.Describe(f, goVersion, libcVersion)
	if err != nil {
		return err
	}

	diffs := localDescriber().Compare(desc)
	if len(diffs) == 0 {
		cmd.Println("No incompatibilities found!")
		return nil
	}

	if gogetEnabled {
		indirects, err := indirectRequires(goSum)
		if err != nil {
			return err
		}
		for _, diff := range diffs {
			if diff.Name != "go" && diff.Name != "libc" {
				if _, ok := indirects[diff.Name]; ok {
					cmd.Printf("go mod edit --replace %s=%s@%s\n", diff.Name, diff.Name, diff.Expected)
				} else {
					cmd.Printf("go get %s@%s\n", diff.Name, diff.Expected)
				}
				continue
			}

			cmd.Println(diff.Name)
			cmd.Println("\thave:", diff.Have)
			cmd.Println("\twant:", diff.Expected)
		}
	} else {
		for _, diff := range diffs {
			cmd.Println(diff.Name)
			cmd.Println("\thave:", diff.Have)
			cmd.Println("\twant:", diff.Expected)
		}
	}

	return fmt.Errorf("%d incompatibilities found", len(diffs))
}
