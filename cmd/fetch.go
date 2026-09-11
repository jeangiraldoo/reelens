package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"reelens/config"
	"reelens/data"
	"reelens/providers"
)

// errFetchFailed signals that at least one package failed to be fetched. Its
// message is never displayed; the per-package failures printed during the
// run are the actual report.
var errFetchFailed = errors.New("fetch failed")

func fetchCmd(cfg config.Config, pkgsData data.LocalPkgs) *cobra.Command {
	return &cobra.Command{
		Use:   "fetch",
		Short: "Updates the local cache for package data",
		RunE: func(cmd *cobra.Command, args []string) error {
			names := cfg.SortedPkgNames()

			if len(args) > 0 {
				names = args
			}

			var failed int
			for _, pkgName := range names {
				pkgConfig, ok := cfg.Pkgs[pkgName]
				if !ok {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not fetch %s: not found in config\n", pkgName)
					failed++
					continue
				}

				err := providers.CacheRelease(pkgName, pkgConfig, pkgsData)
				if err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not fetch %s: %v\n", pkgName, err)
					failed++
				} else {
					fmt.Println("fetched " + pkgName)
				}
			}

			if err := pkgsData.Save(); err != nil {
				return fmt.Errorf("could not save packages: %w", err)
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
}
