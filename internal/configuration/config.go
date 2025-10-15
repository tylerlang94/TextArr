package configuration

import (
	"errors"
	"fmt"
	"strings"
)

type Config struct {
	Sonarr struct {
		URL string `yaml:"url"`
		API string `yaml:"api"`
	} `yaml:"sonarr"`
	Radarr struct {
		URL string `yaml:"url"`
		API string `yaml:"api"`
	} `yaml:"radarr"`
	Paths struct {
		TV     string `yaml:"tvRootPath"`
		Movies string `yaml:"movieRootPath"`
	} `yaml:"paths"`
}

func (c *Config) Validate() error {
	if c.Sonarr.URL == "" || c.Sonarr.API == "" {
		return errors.New("missing Sonarr URL or API key")
	}
	if c.Radarr.URL == "" || c.Radarr.API == "" {
		return errors.New("missing Radarr URL or API key")
	}
	if c.Paths.TV == "" {
		return fmt.Errorf("TV root path is required")
	}
	if c.Paths.Movies == "" {
		return fmt.Errorf("Movie root path is required")
	}
	return nil
}

func (c *Config) Normalize() error {
	// Trim spaces
	c.Sonarr.URL = strings.TrimSpace(c.Sonarr.URL)
	c.Sonarr.API = strings.TrimSpace(c.Sonarr.API)
	c.Radarr.URL = strings.TrimSpace(c.Radarr.URL)
	c.Radarr.API = strings.TrimSpace(c.Radarr.API)
	c.Paths.TV = strings.TrimSpace(c.Paths.TV)
	c.Paths.Movies = strings.TrimSpace(c.Paths.Movies)

	// Remove trailing slash from URLs
	c.Sonarr.URL = strings.TrimRight(c.Sonarr.URL, "/")
	c.Radarr.URL = strings.TrimRight(c.Radarr.URL, "/")

	// Optional: set defaults for paths if not provided
	if c.Paths.TV == "" {
		c.Paths.TV = "/media/tv"
	}
	if c.Paths.Movies == "" {
		c.Paths.Movies = "/media/movies"
	}

	return nil
}
