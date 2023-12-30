// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package config

import (
	"context"
	"errors"
	"os"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v3"
)

// Setup is used to setup the configuration. It will use both the configuration at the given
// path and the environment variables.
func Setup(configPath string) (Config, error) {
	// Read the file with a YAML decoder.
	b, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}
	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return Config{}, err
	}

	// Create the structs if they are unset.
	if config.Path == nil {
		config.Path = &PathConfig{}
	}
	if config.Server == nil {
		config.Server = &ServerConfig{}
	}
	if config.Database == nil {
		config.Database = &DatabaseConfig{}
	}

	// Process the environment variables.
	if err = envconfig.Process(context.Background(), &config); err != nil {
		return Config{}, err
	}

	// Make sure both cert/key file is blank or both are set.
	if config.Server.SSLCertFile != "" || config.Server.SSLKeyFile != "" {
		if config.Server.SSLCertFile == "" || config.Server.SSLKeyFile == "" {
			return Config{}, errors.New("both SSL certificate and key file must be set or both must be blank")
		}
	}

	// Default the host if it is unset.
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0:23452"
	}

	// Return the configuration.
	return config, nil
}
