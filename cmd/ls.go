package cmd

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"reelens/config"
	"reelens/providers"
	"reelens/utils/dates"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists the packages being tracked",

	RunE: func(cmd *cobra.Command, args []string) error {
		table := tablewriter.NewWriter(os.Stdout)

		table.Header("Package", "Latest version", "Published", "Last checked")

		releases, err := providers.LoadReleaseCache()
		if err != nil {
			// Degrade rather than fail: still list tracked packages, minus
			// cached columns. releases is nil here — safe to read, every
			// lookup misses. The write is discarded because losing the
			// warning changes neither the rendered table nor the exit status.
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: showing packages without cached data (%v)\n", err)
		}

		for _, packageName := range config.SortedPackageNames() {
			release := releases[packageName]

			row := []string{
				packageName,
				release.Version,
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
}
