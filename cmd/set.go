package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"reelens/config"
	"reelens/data"
)

func setCmd(config config.Config, localPkgs data.LocalPkgs) *cobra.Command {
	return &cobra.Command{
		Use:   "set <package> <version>",
		Short: "Sets the installed version of a package",
		Long: `Sets the installed version of a package that is already being tracked,
recording it so ls can tell whether the package is outdated.

Example:
  reelens set git v2.62.0`,
		Args: cobra.ExactArgs(2), //nolint:mnd
		RunE: func(cmd *cobra.Command, args []string) error {
			pkgName, version := args[0], args[1]

			if _, err := config.GetPkg(pkgName); err != nil {
				return err
			}

			pkg := localPkgs.Pkgs[pkgName]
			pkg.InstalledVersion = version

			localPkgs.Pkgs[pkgName] = pkg

			if err := localPkgs.Write(); err != nil {
				return fmt.Errorf("could not set %s: %w", pkgName, err)
			}

			fmt.Printf("Set %s to %s\n", pkgName, version)
			return nil
		},
	}
}
