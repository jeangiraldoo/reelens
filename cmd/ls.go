package cmd

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"reelens/config"
	"reelens/providers"
	"reelens/utils/collections"
	"reelens/utils/dates"
)

var (
	outdatedFlag,
	notOutdatedFlag,
	aheadFlag,
	unknownFlag bool
)

func filterPackages() ([]string, error) {
	releases, err := providers.LoadReleaseCache()
	if err != nil {
		// Degrade rather than fail: still list tracked packages, minus
		// cached columns. releases is nil here — safe to read, every
		// lookup misses. The write is discarded because losing the
		// warning changes neither the rendered table nor the exit status.
		return []string{}, err
	}

	pkgs := config.SortedPackageNames()

	if outdatedFlag {
		pkgs = collections.Filter(pkgs, func(element string) bool {
			pkgData := releases[element]

			return pkgData.ResolveVersionStatus() == providers.Outdated
		})
	}

	if notOutdatedFlag {
		pkgs = collections.Filter(pkgs, func(element string) bool {
			pkgData := releases[element]

			return pkgData.ResolveVersionStatus() == providers.Current
		})
	}

	if aheadFlag {
		pkgs = collections.Filter(pkgs, func(element string) bool {
			pkgData := releases[element]

			return pkgData.ResolveVersionStatus() == providers.Ahead
		})
	}

	if unknownFlag {
		pkgs = collections.Filter(pkgs, func(element string) bool {
			pkgData := releases[element]

			return pkgData.ResolveVersionStatus() == providers.Unknown
		})
	}

	return pkgs, nil

}

func resolveOutdatedLabel(release providers.ReleaseCache) (label string) {
	const (
		outdatedLabel = "yes"
		aheadLabel    = "ahead"
		currentLabel  = "current"
		unknownLabel  = "Unknown"
	)

	switch release.ResolveVersionStatus() {
	case providers.Unknown:
		label = unknownLabel
	case providers.Current:
		label = currentLabel
	case providers.Outdated:
		label = outdatedLabel
	case providers.Ahead:
		label = aheadLabel
	}

	return
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists the packages being tracked",

	RunE: func(cmd *cobra.Command, args []string) error {
		table := tablewriter.NewWriter(os.Stdout)

		table.Header("Package", "Latest version", "Installed version", "Outdated", "Published", "Last fetched")

		releases, err := providers.LoadReleaseCache()
		if err != nil {
			// Degrade rather than fail: still list tracked packages, minus
			// cached columns. releases is nil here — safe to read, every
			// lookup misses. The write is discarded because losing the
			// warning changes neither the rendered table nor the exit status.
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: showing packages without cached data (%v)\n", err)
		}

		pkgs, err := filterPackages()

		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: filtering failed (%v)\n", err)
		}

		for _, packageName := range pkgs {
			release := releases[packageName]

			row := []string{
				packageName,
				release.Version,
				release.InstalledVersion,
				resolveOutdatedLabel(release),
				dateUtils.HumanTimeSince(release.PublishedDate),
				dateUtils.HumanTimeSince(release.CachedAt),
			}
			if err := table.Append(row); err != nil {
				return fmt.Errorf("ls: building row for %q: %w", packageName, err)
			}
		}

		if err := table.Render(); err != nil {
			return fmt.Errorf("ls: rendering table: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolVarP(&outdatedFlag, "outdated", "o", false, "filter for outdated packages")
	lsCmd.Flags().BoolVarP(&notOutdatedFlag, "not-outdated", "u", false, "filter for up-to-date packages")
	lsCmd.Flags().BoolVar(&aheadFlag, "ahead", false, "filter for packages ahead of latest")
	lsCmd.Flags().BoolVar(&unknownFlag, "unknown", false, "filter for packages without an installed version")
}
