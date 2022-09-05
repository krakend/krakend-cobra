package cmd

import (
	"github.com/luraproject/lura/v2/core"
	"github.com/spf13/cobra"
)

func versionFunc(cmd *cobra.Command, _ []string) {
	cmd.Println(core.KrakendVersion)
}
