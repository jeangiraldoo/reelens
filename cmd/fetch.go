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
				pkgConfig, err := cfg.GetPkg(pkgName)
				if err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not fetch %s: %v\n", pkgName, err)
					failed++
					continue
				}

				provider, err := providers.Lookup(pkgConfig.Provider.Type)
				if err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not fetch %s: %v\n", pkgName, err)
					failed++
					continue
				}

				release, err := provider.GetLatestRelease(pkgName, pkgConfig)
				if err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not fetch %s: %v\n", pkgName, err)
					failed++
					continue
				}

				pkgsData.SetRelease(pkgName, release)
				if pkgConfig.Changelog != "" {
					err = fetchChangelog(pkgConfig, pkgConfig.Changelog, pkgName)
					if err != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not fetch the changelog for %s: %v\n", pkgName, err)
						failed++
						continue
					}

				}
				fmt.Println("fetched " + pkgName)
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

func fetchChangelog(pkg config.Pkg, changelogFileName, pkgName string) error {
	provider, err := providers.Lookup(pkg.Provider.Type)
	if err != nil {
		return err
	}

	cfg, err := providers.DecodeProviderConfig[struct {
		RepoID string `yaml:"repoID"`
	}](pkg)
	if err != nil {
		return err
	}

	changelog, err := provider.GetFile(cfg.RepoID, changelogFileName)
	if err != nil {
		return err
	}

	return data.Changelogs{}.Write(pkgName, changelog)
}
