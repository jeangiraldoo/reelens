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
	Packages map[string]Package `yaml:"packages"`
}

type Package struct {
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

var Cfg Config

// DefaultPath returns the location reelens reads its configuration from.
func DefaultPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "reelens", "config.yaml")
}

// Load reads and parses the configuration file into Cfg. A missing config
// file is not an error: Cfg simply stays empty.
func Load() error {
	path := DefaultPath()
	if path == "" {
		return errors.New("could not determine user config directory")
	}

	data, err := os.ReadFile(path)
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

func SortedPackageNames() []string {
	names := make([]string, 0, len(Cfg.Packages))
	for name := range Cfg.Packages {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
