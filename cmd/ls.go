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

func filterPackages() ([]string, error) {
	releases, err := data.LoadReleaseCache()
	if err != nil {
		// Degrade rather than fail: still list tracked packages, minus
		// cached columns. releases is nil here — safe to read, every
		// lookup misses. The write is discarded because losing the
		// warning changes neither the rendered table nor the exit status.
		return []string{}, err
	}

	pkgs := config.SortedPackageNames()

	if allFlag {
		return pkgs, nil
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

	pkgs = collections.Filter(pkgs, func(element string) bool {
		status := releases[element].ResolveVersionStatus()
		for _, filter := range filters {
			if filter(status) {
				return true
			}
		}
		return false
	})

	return pkgs, nil

}

func resolveOutdatedLabel(release data.LocalPackageData) (label string) {

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

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists the packages being tracked",

	RunE: func(cmd *cobra.Command, args []string) error {

		releases, err := data.LoadReleaseCache()
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

		if len(pkgs) == 0 {
			fmt.Println("no new versions available")
			return nil
		}

		if longFlag {
			if err = longOutput(pkgs, releases); err != nil {
				return err
			}

			return nil
		}

		nameWidth := collections.LongestString(pkgs)

		versionWidth := 0
		for _, packageName := range pkgs {
			currentVersion := releases[packageName].InstalledVersion
			if currentVersion == "" {
				currentVersion = unknownLabel
			}
			if len(currentVersion) > versionWidth {
				versionWidth = len(currentVersion)
			}
		}

		for _, packageName := range pkgs {
			release := releases[packageName]
			currentVersion := release.InstalledVersion

			if currentVersion == "" {
				currentVersion = unknownLabel
			}

			fmt.Printf("%-*s %-*s -> %s\n", nameWidth, packageName, versionWidth, currentVersion, release.Version)
		}

		return nil
	},
}

func longOutput(pkgs []string, releases map[string]data.LocalPackageData) error {
	table := tablewriter.NewWriter(os.Stdout)

	table.Header("Package", "Latest version", "Installed version", "Outdated", "Published", "Last fetched")

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
}

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolVarP(&notOutdatedFlag, "not-outdated", "u", false, "filter for up-to-date packages")
	lsCmd.Flags().BoolVar(&aheadFlag, "ahead", false, "filter for packages ahead of latest")
	lsCmd.Flags().BoolVar(&unknownFlag, "unknown", false, "filter for packages without an installed version")
	lsCmd.Flags().BoolVarP(&longFlag, "long", "l", false, "show detailed table output")
	lsCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "show all packages")
}
