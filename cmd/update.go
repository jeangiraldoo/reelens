package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"reelens/config"
	"reelens/providers"
)

// errUpdateFailed signals that at least one package failed to update. Its
// message is never displayed; the per-package failures printed during the
// run are the actual report.
var errUpdateFailed = errors.New("update failed")

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates the local cache for package data",
	RunE: func(cmd *cobra.Command, args []string) error {
		names := config.SortedPackageNames()

		if len(args) > 0 {
			names = args
		}

		var failed int
		for _, packageName := range names {
			pkgConfig, ok := config.Cfg.Packages[packageName]
			if !ok {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not update %s: not found in config\n", packageName)
				failed++
				continue
			}

			err := providers.CacheRelease(packageName, pkgConfig)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not update %s: %v\n", packageName, err)
				failed++
			} else {
				fmt.Println("Updated " + packageName)
			}
		}

		if failed > 0 {
			// Reasons were already reported per package; this error exists
			// only to drive a nonzero exit code.
			cmd.SilenceErrors = true
			return errUpdateFailed
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
