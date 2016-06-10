package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// secretsCmd represents the secrets command
var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage secrets & keys",
	Long:  `Manage secrets & keys`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("syntax: create describe expose get")
	},
}

func init() {
	rootCommand.AddCommand(secretsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// secretsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// secretsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
