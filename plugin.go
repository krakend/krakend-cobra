package cmd

import (
	"os"

	"github.com/devopsfaith/krakend-cobra/v2/plugin"
	"github.com/spf13/cobra"
)

func pluginFunc(cmd *cobra.Command, args []string) {
	f, err := os.Open(goSum)
	if err != nil {
		cmd.Println(err)
		os.Exit(1)
		return
	}
	defer f.Close()

	desc, err := plugin.Describe(f)
	if err != nil {
		cmd.Println(err)
		os.Exit(1)
		return
	}
	desc.Go = goVersion

	diffs := plugin.Compare(plugin.Local(), desc)

	if len(diffs) == 0 {
		cmd.Println("No incompatibilities found!")
		return
	}

	cmd.Println(len(diffs), "incompatibility(ies) found...")
	for _, diff := range diffs {
		cmd.Println(diff.Name)
		cmd.Println("\thave:", diff.Have)
		cmd.Println("\twant:", diff.Expected)
	}
	os.Exit(1)
}
