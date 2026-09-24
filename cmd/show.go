package cmd

import (
	"errors"
	"github.com/charmbracelet/gum/pager"
	"github.com/spf13/cobra"
	"reelens/config"
	"reelens/providers"
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

			pkg, err := cfg.GetPkg(pkgName)
			if err != nil {
				return err
			}

			switch resource {
			case "changelog":
				err := showChangelog(pkg, pkg.Changelog)
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

func showChangelog(pkg config.Pkg, changelogFileName string) error {
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
