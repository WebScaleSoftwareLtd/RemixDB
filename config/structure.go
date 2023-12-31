// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package config

// PathConfig is used to define the path configuration structure.
type PathConfig struct {
	// Data defines the data path.
	Data string `yaml:"data" env:"REMIXDB_DATA_PATH,overwrite"`

	// GoPlugin defines the Go plugin path.
	GoPlugin string `yaml:"go_plugin" env:"REMIXDB_GOPLUGIN_PATH,overwrite"`
}

// DatabaseConfig is used to define the database configuration structure.
type DatabaseConfig struct {
	// PartitionsEnabled defines if partitions are enabled.
	PartitionsEnabled bool `yaml:"partitions_enabled" env:"PARTITIONS_ENABLED,overwrite"`
}

// ServerConfig is used to define the server configuration structure.
type ServerConfig struct {
	// SSLCertFile defines the SSL certificate file.
	SSLCertFile string `yaml:"ssl_cert_file" env:"SSL_CERT_FILE,overwrite"`

	// SSLKeyFile defines the SSL key file.
	SSLKeyFile string `yaml:"ssl_key_file" env:"SSL_KEY_FILE,overwrite"`

	// H2C defines if H2C is enabled.
	H2C bool `yaml:"h2c" env:"H2C,overwrite"`

	// Host defines the host to listen on.
	Host string `yaml:"host" env:"HOST,overwrite"`

	// XForwardedHost defines if the X-Forwarded-Host header should be used.
	XForwardedHost bool `yaml:"x_forwarded_host" env:"X_FORWARDED_HOST,overwrite"`
}

// Config is used to define the main configuration structure.
type Config struct {
	// Debug defines if debug mode is enabled.
	Debug bool `yaml:"debug" env:"DEBUG,overwrite"`

	// Path defines the path configuration.
	Path *PathConfig `yaml:"path"`

	// Database defines the database configuration.
	Database *DatabaseConfig `yaml:"database"`

	// Server defines the server configuration.
	Server *ServerConfig `yaml:"server"`
}
