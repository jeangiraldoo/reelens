package cmd

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"reelens/config"
	"reelens/data"
	"reelens/utils/collections"
	"reelens/utils/dates"
)

var (
	allFlag,
	notOutdatedFlag,
	aheadFlag,
	unknownFlag,
	longFlag bool
)

const (
	outdatedLabel = "yes"
	aheadLabel    = "ahead"
	currentLabel  = "current"
	unknownLabel  = "Unknown"
)

// filterPkgs applies the requested status filters to the configured package
// names.
func filterPkgs(pkgs []string, pkgsData data.LocalPkgs) []string {
	if allFlag {
		return pkgs
	}

	var filters []func(data.ReleaseState) bool

	if notOutdatedFlag {
		filters = append(filters, func(status data.ReleaseState) bool {
			return status == data.Current
		})
	}

	if aheadFlag {
		filters = append(filters, func(status data.ReleaseState) bool {
			return status == data.Ahead
		})
	}

	if unknownFlag {
		filters = append(filters, func(status data.ReleaseState) bool {
			return status == data.Unknown
		})
	}

	if len(filters) == 0 {
		// Default to only outdated packages when no status filter was requested.
		filters = append(filters, func(status data.ReleaseState) bool {
			return status == data.Outdated
		})
	}

	return collections.Filter(pkgs, func(element string) bool {
		status := pkgsData[element].ResolveVersionStatus()
		for _, filter := range filters {
			if filter(status) {
				return true
			}
		}
		return false
	})
}

func resolveOutdatedLabel(release data.LocalPkg) (label string) {

	switch release.ResolveVersionStatus() {
	case data.Unknown:
		label = unknownLabel
	case data.Current:
		label = currentLabel
	case data.Outdated:
		label = outdatedLabel
	case data.Ahead:
		label = aheadLabel
	}

	return
}

func lsCmd(cfg config.Config, pkgsData data.LocalPkgs) *cobra.Command {
	var lsCmd = &cobra.Command{
		Use:   "ls",
		Short: "Lists the packages being tracked",

		RunE: func(cmd *cobra.Command, args []string) error {
			pkgs := filterPkgs(cfg.SortedPkgNames(), pkgsData)

			if len(pkgs) == 0 {
				fmt.Println("no new versions available")
				return nil
			}

			if longFlag {
				return longOutput(pkgs, pkgsData)
			}

			nameWidth := collections.LongestString(pkgs)

			versionWidth := 0
			for _, pkgName := range pkgs {
				currentVersion := pkgsData[pkgName].InstalledVersion
				if currentVersion == "" {
					currentVersion = unknownLabel
				}
				if len(currentVersion) > versionWidth {
					versionWidth = len(currentVersion)
				}
			}

			for _, pkgName := range pkgs {
				release := pkgsData[pkgName]
				currentVersion := release.InstalledVersion

				if currentVersion == "" {
					currentVersion = unknownLabel
				}

				fmt.Printf("%-*s %-*s -> %s\n", nameWidth, pkgName, versionWidth, currentVersion, release.LatestVersion)
			}

			return nil
		},
	}

	lsCmd.Flags().BoolVarP(&notOutdatedFlag, "not-outdated", "u", false, "filter for up-to-date packages")
	lsCmd.Flags().BoolVar(&aheadFlag, "ahead", false, "filter for packages ahead of latest")
	lsCmd.Flags().BoolVar(&unknownFlag, "unknown", false, "filter for packages without an installed version")
	lsCmd.Flags().BoolVarP(&longFlag, "long", "l", false, "show detailed table output")
	lsCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "show all packages")

	return lsCmd
}

func longOutput(pkgs []string, localPkgsData data.LocalPkgs) error {
	table := tablewriter.NewWriter(os.Stdout)

	table.Header("Package", "Installed version", "Latest version", "Outdated", "Published", "Last fetched")

	for _, pkgName := range pkgs {
		localPkgData := localPkgsData[pkgName]

		row := []string{
			pkgName,
			localPkgData.InstalledVersion,
			localPkgData.LatestVersion,
			resolveOutdatedLabel(localPkgData),
			dateUtils.HumanTimeSince(localPkgData.PublishedDate),
			dateUtils.HumanTimeSince(localPkgData.FetchedAt),
		}
		if err := table.Append(row); err != nil {
			return fmt.Errorf("ls: building row for %q: %w", pkgName, err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("ls: rendering table: %w", err)
	}

	return nil
}
