package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reelens/appinfo"
	"reelens/utils/system"
	"sort"
)

type Config struct {
	Pkgs map[string]Pkg `yaml:"packages"`
}

type Pkg struct {
	Provider  ProviderRef `yaml:"provider"`
	Version   string      `yaml:"version"`
	Changelog string      `yaml:"changelog"`
}

// ProviderRef holds the provider type and its raw YAML node. The registered
// provider decodes the node into its own configuration.
type ProviderRef struct {
	Type string
	Node *yaml.Node
}

func (p *ProviderRef) UnmarshalYAML(node *yaml.Node) error {
	p.Node = node

	var head struct {
		Type string `yaml:"type"`
	}
	if err := node.Decode(&head); err != nil {
		return err
	}
	p.Type = head.Type
	return nil
}

const (
	configDirName      = appinfo.Name
	mainConfigFileName = "config.yaml"
)

// Load reads and parses the configuration file. A missing config file is not
// an error: an empty Config is returned.
func Load() (Config, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, errors.New("could not determine user config directory")
	}

	mainConfigFilePath := filepath.Join(userConfigDir, configDirName, mainConfigFileName)

	data, err := system.ReadFile(mainConfigFilePath)
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("cannot parse the config file: %w", err)
	}

	return cfg, nil
}

func (c Config) SortedPkgNames() []string {
	names := make([]string, 0, len(c.Pkgs))
	for name := range c.Pkgs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetPkg returns the config entry for pkgName.
func (c Config) GetPkg(pkgName string) (Pkg, error) {
	pkgConfig, ok := c.Pkgs[pkgName]
	if !ok {
		return Pkg{}, fmt.Errorf("package %q not found in config", pkgName)
	}

	return pkgConfig, nil
}
