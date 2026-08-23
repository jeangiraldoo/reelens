package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"reelens/config"
	"reelens/providers"
	_ "reelens/providers/all"
)

var rootCmd = &cobra.Command{
	Use:   "reelens",
	Short: "Tracks package changes, commits and releases.",
	Long:  "Tracks package changes, commits and releases so you don't have to do it yourself.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.Root().SilenceUsage = true

		if err := config.Load(); err != nil {
			return err
		}
		return providers.Init()
	},
}

// Execute is called by main.main(). It only needs to happen once to rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
