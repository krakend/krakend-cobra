package cmd

import (
	"bytes"
	"testing"

	"github.com/krakendio/krakend-cobra/v2/plugin"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func Test_pluginFuncErr(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOutput(&buf)

	localDescriber = func() plugin.Descriptor {
		return plugin.Descriptor{
			Go:   goVersion,
			Libc: libcVersion,
			Deps: map[string]string{
				"golang.org/x/mod":                  "v0.6.0-dev.0.20220419223038-86c51ed26bb4",
				"github.com/Azure/azure-sdk-for-go": "v59.3.0+incompatible",
				"cloud.google.com/go":               "v0.100.2",
			},
		}
	}

	defer func() { localDescriber = plugin.Local }()

	tests := map[string]struct {
		goSum    string
		expected string
		fix      bool
		err      string
	}{
		"missing": {
			goSum: "./testdata/missing-go.sum",
			err:   "open ./testdata/missing-go.sum: no such file or directory",
		},
		"matching": {
			goSum:    "./testdata/match-go.sum",
			expected: "No incompatibilities found!\n",
		},
		"changes": {
			goSum: "./testdata/changes-go.sum",
			expected: `cloud.google.com/go
	have: v0.100.3
	want: v0.100.2
github.com/Azure/azure-sdk-for-go
	have: v59.3.1+incompatible
	want: v59.3.0+incompatible
golang.org/x/mod
	have: v0.6.10-dev.0.20220419223038-86c51ed26bb4
	want: v0.6.0-dev.0.20220419223038-86c51ed26bb4
`,
			err: "3 incompatibilities found",
		},
		"fix": {
			goSum: "./testdata/changes-go.sum",
			fix:   true,
			expected: `go mod edit --replace cloud.google.com/go=cloud.google.com/go@v0.100.2
go mod edit --replace github.com/Azure/azure-sdk-for-go=github.com/Azure/azure-sdk-for-go@v59.3.0+incompatible
go get golang.org/x/mod@v0.6.0-dev.0.20220419223038-86c51ed26bb4
`,
			err: "3 incompatibilities found",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			buf.Reset()

			orig := goSum
			goSum = tc.goSum
			defer func() { goSum = orig }()

			fix := gogetEnabled
			gogetEnabled = tc.fix
			defer func() { gogetEnabled = fix }()

			err := pluginFuncErr(cmd, nil)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expected, buf.String())
		})
	}
}
