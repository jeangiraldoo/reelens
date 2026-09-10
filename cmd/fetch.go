package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"reelens/config"
	"reelens/providers"
)

// errFetchFailed signals that at least one package failed to be fetched. Its
// message is never displayed; the per-package failures printed during the
// run are the actual report.
var errFetchFailed = errors.New("fetch failed")

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Updates the local cache for package data",
	RunE: func(cmd *cobra.Command, args []string) error {
		names := config.SortedPkgNames()

		if len(args) > 0 {
			names = args
		}

		var failed int
		for _, pkgName := range names {
			pkgConfig, ok := config.Cfg.Pkgs[pkgName]
			if !ok {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not fetch %s: not found in config\n", pkgName)
				failed++
				continue
			}

			err := providers.CacheRelease(pkgName, pkgConfig)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not fetch %s: %v\n", pkgName, err)
				failed++
			} else {
				fmt.Println("fetched " + pkgName)
			}
		}

		if failed > 0 {
			// Reasons were already reported per package; this error exists
			// only to drive a nonzero exit code.
			cmd.SilenceErrors = true
			return errFetchFailed
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
