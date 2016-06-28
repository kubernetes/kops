package main

import (
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "import clusters",
	Long:  `import clusters`,
}

func init() {
	rootCommand.AddCommand(importCmd)
}
