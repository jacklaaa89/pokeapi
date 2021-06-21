package cmd

import "github.com/spf13/cobra"

// rootCmd is the main command to execute.
var rootCmd = &cobra.Command{
	Use:   "pokeapi",
	Short: "An API which queries pokemon data",
	Long: "An API which enables you to query for information on pokemon using " +
		"the pokeapi, also allowing for translations",
}

func init() { rootCmd.AddCommand(serveCmd) }

// Root returns the root command.
func Root() *cobra.Command { return rootCmd }
