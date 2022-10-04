package cmd

import (
	"os"

	"github.com/krakendio/krakend-cobra/v2/plugin"
	"github.com/spf13/cobra"
)

func pluginFunc(cmd *cobra.Command, _ []string) {
	f, err := os.Open(goSum)
	if err != nil {
		cmd.Println(err)
		os.Exit(1)
		return
	}

	desc, err := plugin.Describe(f, goVersion, libcVersion)
	if err != nil {
		cmd.Println(err)
		f.Close()
		os.Exit(1)
		return
	}

	diffs := plugin.Local().Compare(desc)
	if len(diffs) == 0 {
		cmd.Println("No incompatibilities found!")
		f.Close()
		return
	}

	cmd.Println(len(diffs), "incompatibility(ies) found...")
	if gogetEnabled {
		for _, diff := range diffs {
			if diff.Name != "go" && diff.Name != "libc" {
				cmd.Printf("go get %s@%s\n", diff.Name, diff.Expected)
				continue
			}

			cmd.Println(diff.Name)
			cmd.Println("\thave:", diff.Have)
			cmd.Println("\twant:", diff.Expected)
		}
		f.Close()
		os.Exit(1)
	}

	for _, diff := range diffs {
		cmd.Println(diff.Name)
		cmd.Println("\thave:", diff.Have)
		cmd.Println("\twant:", diff.Expected)
	}
	f.Close()
	os.Exit(1)
}
