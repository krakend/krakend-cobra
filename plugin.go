package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/krakendio/krakend-cobra/v2/plugin"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

const (
	// goName is the name of the diff which represents the go version.
	goName = "go"

	// libcName is the name of the diff which represents the libc version.
	libcName = "libc"
)

// goMod returns the go.mod file path from the go.sum file path.
func goMod(goSum string) string {
	return filepath.Join(filepath.Dir(goSum), "go.mod")
}

// indirectRequires returns the details and indirect dependencies of the go.sum file.
func indirectRequires(goSum string) (*modfile.File, map[string]struct{}, error) {
	filename := goMod(goSum)
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("read go.mod: %w", err)
	}

	f, err := modfile.Parse(filename, data, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("parse go.mod: %w", err)
	}

	indirects := map[string]struct{}{}
	for _, r := range f.Require {
		if r.Indirect {
			indirects[r.Mod.Path] = struct{}{}
		}
	}

	return f, indirects, nil
}

// writeModFile writes the modfile.File to the go.mod file determined from goSum.
func writeModFile(goSum string, f *modfile.File) error {
	f.Cleanup()
	data, err := f.Format()
	if err != nil {
		return fmt.Errorf("format go.mod: %w", err)
	}

	filename := goMod(goSum)
	if err = os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("write go.sum: %w", err)
	}

	return nil
}

// getBuildInfo returns the dependencies of the binary calling it.
// It is a var to allow the replacement of the function in the tests
// as the debug.ReadBuildInfo function is not available in the tests
// https://github.com/golang/go/issues/68045
var localDescriber = plugin.Local

func pluginFunc(cmd *cobra.Command, args []string) {
	if err := pluginFuncErr(cmd, args); err != nil {
		cmd.Println(err)
		os.Exit(1)
	}
}

func pluginFuncErr(cmd *cobra.Command, _ []string) error {
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

	var indirects map[string]struct{}
	var modFile *modfile.File
	if gogetEnabled || fixEnabled {
		if modFile, indirects, err = indirectRequires(goSum); err != nil {
			return err
		}
	}

	var fixed int
	if !fixEnabled {
		outputFixes(cmd, diffs, indirects)
	} else if fixed, err = applyFixes(cmd, diffs, modFile, indirects); err != nil {
		return err
	}

	// Report any remaining incompatibilities.
	if len(diffs) != fixed {
		if fixed > 0 {
			return fmt.Errorf("%d incompatibilities fixed, %d left", fixed, len(diffs)-fixed)
		}

		return fmt.Errorf("%d incompatibilities found", len(diffs))
	}

	return nil
}

// outputFixes prints the incompatibilities.
func outputFixes(cmd *cobra.Command, diffs []plugin.Diff, indirects map[string]struct{}) {
	for _, diff := range diffs {
		if diff.Name != goName && diff.Name != libcName && gogetEnabled {
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
}

// applyFixes applies the fixes and returns the number of incompatibilities fixed.
func applyFixes(cmd *cobra.Command, diffs []plugin.Diff, modFile *modfile.File, indirects map[string]struct{}) (int, error) {
	var replaces []plugin.Diff
	var requires []*modfile.Require
	for _, diff := range diffs {
		if diff.Name != goName && diff.Name != libcName {
			if _, ok := indirects[diff.Name]; ok {
				replaces = append(replaces, diff)
				continue
			}

			requires = append(requires, &modfile.Require{
				Mod: module.Version{
					Path:    diff.Name,
					Version: diff.Expected,
				},
			})
			continue
		}

		cmd.Println(diff.Name)
		cmd.Println("\thave:", diff.Have)
		cmd.Println("\twant:", diff.Expected)
	}

	if len(requires) > 0 {
		// We use the modfile.SetRequireSeparateIndirect to avoid adding direct
		// dependencies to the direct block as per:
		// https://github.com/golang/go/issues/69050
		modFile.SetRequireSeparateIndirect(append(modFile.Require, requires...))
	}

	// Add replaces after requires.
	for _, diff := range replaces {
		if err := modFile.AddReplace(diff.Name, "", diff.Name, diff.Expected); err != nil {
			return 0, fmt.Errorf("add replace: %w", err)
		}
	}

	fixed := len(replaces) + len(requires)
	if fixed > 0 {
		modFile.Cleanup()
		if err := writeModFile(goSum, modFile); err != nil {
			return 0, err
		}

		cmd.Printf("%d incompatibilities fixed\n", fixed)
	}

	return fixed, nil
}
