package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sort"
)

type Config struct {
	Pkgs map[string]Pkg `yaml:"packages"`
}

type Pkg struct {
	Provider  ProviderRef `yaml:"provider"`
	Version   string      `yaml:"version"`
	Changelog Changelog   `yaml:"changelog"`
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

type Changelog struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

const (
	configDirName      = "reelens"
	mainConfigFileName = "config.yaml"
)

var Cfg Config

// Load reads and parses the configuration file into Cfg. A missing config
// file is not an error: Cfg simply stays empty.
func Load() error {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return errors.New("could not determine user config directory")
	}

	mainConfigFilePath := filepath.Join(userConfigDir, configDirName, mainConfigFileName)

	data, err := os.ReadFile(mainConfigFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // first run, no config yet
		}
		return fmt.Errorf("cannot read the config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		return fmt.Errorf("cannot parse the config file: %w", err)
	}

	return nil
}

func SortedPkgNames() []string {
	names := make([]string, 0, len(Cfg.Pkgs))
	for name := range Cfg.Pkgs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
