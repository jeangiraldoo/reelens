package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"reelens/appinfo"
	"reelens/config"
	"reelens/data"
	_ "reelens/providers/all"
)

func NewRoot(cfg config.Config) (*cobra.Command, error) {
	root := &cobra.Command{
		Use:          appinfo.Name,
		Version:      appinfo.Version(),
		Short:        "Tracks package changes, commits and releases.",
		Long:         "Tracks package changes, commits and releases so you don't have to do it yourself.",
		SilenceUsage: true,
	}

	localPkgs, err := data.NewLocalPkgs()
	if err != nil {
		return nil, fmt.Errorf("could not load packages: %w", err)
	}

	root.AddCommand(fetchCmd(cfg, localPkgs))
	root.AddCommand(lsCmd(cfg, localPkgs))
	root.AddCommand(setCmd(cfg, localPkgs))
	root.AddCommand(showCmd(cfg))

	return root, nil
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

	command, err := NewRoot(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
