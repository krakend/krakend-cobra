package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krakendio/krakend-cobra/v2/plugin"
	"github.com/luraproject/lura/v2/core"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

const testDir = "testdata"

// copyDir is a helper function to copy directory entries from src to dst.
func copyDir(t *testing.T, srcSubDir, dstDir string) {
	t.Helper()

	srcDir := filepath.Join(testDir, srcSubDir)
	entries, err := os.ReadDir(srcDir)
	if errors.Is(err, os.ErrNotExist) {
		// Nothing to do.
		return
	}
	require.NoError(t, err)

	for _, entry := range entries {
		file := entry.Name()
		data, err := os.ReadFile(filepath.Join(srcDir, file))
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(dstDir, file), data, 0644)
		require.NoError(t, err)
	}
}

// loadFile is a helper function to load a file from the testdata directory.
func loadFile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(testDir, name))
	require.NoError(t, err)

	return string(data)
}

func Test_pluginFuncErr(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOutput(&buf)

	localDescriber = func() plugin.Descriptor {
		return plugin.Descriptor{
			Go:   core.GoVersion,
			Libc: core.GlibcVersion,
			Deps: map[string]string{
				"golang.org/x/mod":                  "v0.6.0-dev.0.20220419223038-86c51ed26bb4",
				"github.com/Azure/azure-sdk-for-go": "v59.3.0+incompatible",
				"cloud.google.com/go":               "v0.100.2",
			},
		}
	}

	defer func() { localDescriber = plugin.Local }()

	goModData := loadFile(t, "go.mod")

	tests := map[string]struct {
		dir           string
		expected      string
		expectedGoMod string
		goVersion     string
		format        bool
		fix           bool
		err           string
	}{

		"missing": {
			dir:           "missing",
			goVersion:     goVersion,
			expectedGoMod: goModData,
			err:           "open DIR/go.sum: no such file or directory",
		},

		"match": {
			dir:           "match",
			goVersion:     goVersion,
			expectedGoMod: goModData,
			expected:      "No incompatibilities found!\n",
		},
		"changes": {
			dir:           "changes",
			goVersion:     goVersion,
			expectedGoMod: goModData,
			expected:      loadFile(t, "changes/expected.txt"),
			err:           "3 incompatibilities found",
		},
		"format": {
			dir:           "changes",
			goVersion:     goVersion,
			expectedGoMod: goModData,
			format:        true,
			expected:      loadFile(t, "format/expected.txt"),
			err:           "3 incompatibilities found",
		},
		"fixed-all": {
			dir:           "changes",
			goVersion:     goVersion,
			expectedGoMod: loadFile(t, "fixed-all/go.mod"),
			fix:           true,
			expected:      "3 incompatibilities fixed\n",
		},
		"fixed-some": {
			dir:           "changes",
			goVersion:     "1.1.0",
			expectedGoMod: loadFile(t, "fixed-some/go.mod"),
			fix:           true,
			expected: `go
	have: 1.1.0
	want: undefined
3 incompatibilities fixed
`,
			err: "3 incompatibilities fixed, 1 left",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			buf.Reset()

			// Make copies in a temporary directory so
			// the original files are not modified.
			tempDir := t.TempDir()
			orig := goSum
			goSum = filepath.Join(tempDir, "go.sum")
			copyDir(t, tc.dir, tempDir)
			defer func() { goSum = orig }()

			// Override the global variables for the test.
			format := gogetEnabled
			fix := fixEnabled
			gogetEnabled = tc.format
			fixEnabled = tc.fix
			oldGoVersion := goVersion
			goVersion = tc.goVersion
			defer func() {
				gogetEnabled = format
				fixEnabled = fix
				goVersion = oldGoVersion
			}()

			err := pluginFuncErr(cmd, nil)
			if tc.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tc.err, "DIR", tempDir))
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expected, buf.String())

			data, err := os.ReadFile(filepath.Join(tempDir, "go.mod"))
			if errors.Is(err, os.ErrNotExist) {
				return
			}
			require.NoError(t, err)
			require.Equal(t, string(tc.expectedGoMod), string(data))
		})
	}
}
