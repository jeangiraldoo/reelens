package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"reelens/config"
	"reelens/data"
	_ "reelens/providers/all"
)

func NewRoot(cfg config.Config) *cobra.Command {
	root := &cobra.Command{
		Use:          "reelens",
		Short:        "Tracks package changes, commits and releases.",
		Long:         "Tracks package changes, commits and releases so you don't have to do it yourself.",
		SilenceUsage: true,
	}

	root.AddCommand(fetchCmd(cfg))
	root.AddCommand(lsCmd(cfg))
	root.AddCommand(setCmd())

	return root
}

// Execute is called by main.main(). It loads the configuration and the data
// store, builds the command tree from them, and runs it. It only needs to
// happen once.
func Execute() {
	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := data.Init(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := NewRoot(cfg).Execute(); err != nil {
		os.Exit(1)
	}
}