package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	KnownPartners []string              `yaml:"known_partners"`
	Categories    map[string]*Category  `yaml:"categories"`
}

// Category represents a transaction category with matching rules
type Category struct {
	SubCategory   string   `yaml:"subcategory,omitempty"`
	Keywords      []string `yaml:"keywords"`
	ExactMatches  []string `yaml:"exact_matches,omitempty"`
}

// LoadConfig reads and parses the YAML configuration file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Categories == nil {
		config.Categories = make(map[string]*Category)
	}

	return &config, nil
}

// SaveConfig writes the configuration to a YAML file
func SaveConfig(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// IsKnownPartner checks if a partner is in the known partners list
func (c *Config) IsKnownPartner(partner string) bool {
	for _, known := range c.KnownPartners {
		if known == partner {
			return true
		}
	}
	return false
}

// AddKnownPartner adds a partner to the known partners list if not already present
func (c *Config) AddKnownPartner(partner string) {
	if !c.IsKnownPartner(partner) {
		c.KnownPartners = append(c.KnownPartners, partner)
	}
}
