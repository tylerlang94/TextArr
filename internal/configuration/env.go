package configuration

import "os"

// ApplyEnv overlays environment variables on top of any loaded YAML config.
// This is ideal for secrets or overriding in Docker.
func (c *Config) ApplyEnv() {
	if v := os.Getenv("SONARR_URL"); v != "" {
		c.Sonarr.URL = v
	}
	if v := os.Getenv("SONARR_API"); v != "" {
		c.Sonarr.API = v
	}

	if v := os.Getenv("RADARR_URL"); v != "" {
		c.Radarr.URL = v
	}
	if v := os.Getenv("RADARR_API"); v != "" {
		c.Radarr.API = v
	}

	if v := os.Getenv("TV_ROOT_PATH"); v != "" {
		c.Paths.TV = v
	}
	if v := os.Getenv("MOVIE_ROOT_PATH"); v != "" {
		c.Paths.Movies = v
	}
}
