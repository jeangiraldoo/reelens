package cmd

import (
	"errors"
	"github.com/charmbracelet/gum/pager"
	"github.com/spf13/cobra"
	"reelens/config"
	"reelens/data"
)

func showCmd(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "show <resource> <package>",
		Short: "Shows information about a tracked package",
		Long: `Shows a piece of information about a tracked package.

The first argument selects what to show, for example a changelog:

  reelens show changelog neovim`,
		Args: cobra.ExactArgs(2), //nolint:mnd
		RunE: func(cmd *cobra.Command, args []string) error {
			resource, pkgName := args[0], args[1]
			if _, err := cfg.GetPkg(pkgName); err != nil {
				return err
			}

			switch resource {
			case "changelog":
				err := showChangelog(pkgName)
				if err != nil {
					return err
				}
			default:
				return errors.New("unknown type: " + resource)
			}

			return nil
		},
	}
}

func showChangelog(pkgName string) error {
	changelog, err := data.Changelogs{}.Read(pkgName)
	if err != nil {
		return err
	}

	err = pager.Options{
		Content:         string(changelog),
		ShowLineNumbers: true,
		SoftWrap:        true,
	}.Run()

	if err != nil {
		return err
	}

	return nil
}
