/*
Copyright 2018 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

const (
	V3ioConfigEnvironmentVariable = "V3IO_BACKUP_CONFIG"
	DefaultConfigurationFileName  = "v3io-backup-config.yaml"

	DefaultLogLevel = "info"

	defaultScannerParallelism = 16
	defaultTimeoutInSeconds   = 24 * 60 * 60 // 24 hours
)

type BuildInfo struct {
	BuildTime    string `json:"buildTime,omitempty"`
	Os           string `json:"os,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	Version      string `json:"version,omitempty"`
	CommitHash   string `json:"commitHash,omitempty"`
	Branch       string `json:"branch,omitempty"`
}

func (bi *BuildInfo) String() string {
	return fmt.Sprintf("Build time: %s\nOS: %s\nArchitecture: %s\nVersion: %s\nCommit Hash: %s\nBranch: %s\n",
		bi.BuildTime,
		bi.Os,
		bi.Architecture,
		bi.Version,
		bi.CommitHash,
		bi.Branch)
}

var (
	// Note, following variables set by make
	buildTime, osys, architecture, version, commitHash, branch string

	instance *Config
	once     sync.Once
	failure  error

	BuildMetadta = &BuildInfo{
		BuildTime:    buildTime,
		Os:           osys,
		Architecture: architecture,
		Version:      version,
		CommitHash:   commitHash,
		Branch:       branch,
	}
)

func Error() error {
	return failure
}

type Config struct {
	// V3IO connection information - web-gateway service endpoint,
	// The data container, relative path within the container, and
	// authentication credentials for the web-gateway service
	WebApiEndpoint string `json:"webApiEndpoint"`
	Container      string `json:"container"`
	Path           string `json:"path"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password,omitempty"`
	AccessKey      string `json:"accessKey,omitempty"`

	HttpTimeout string `json:"httpTimeout,omitempty"`
	// Log level - "debug" | "info" | "warn" | "error"
	LogLevel string `json:"logLevel,omitempty"`
	// Number of scan workers (Scan API provides part out of parts interface)
	ScannerParallelism int `json:"scannerParallelism"`
	// Default timeout duration, in seconds; default = 3,600 seconds (1 hour)
	DefaultTimeoutInSeconds int `json:"defaultTimeoutInSeconds,omitempty"`

	// Desired size of single index file
	IndexFileSizeLimit int `json:"indexFileSizeLimit,omitempty"`
	// Desired size of single pack file
	PackFileSizeLimit int `json:"packFileSizeLimit,omitempty"`
	// Metrics-reporter configuration
	MetricsReporter MetricsReporterConfig `json:"performance,omitempty"`
	// Build Info
	BuildInfo *BuildInfo `json:"buildInfo,omitempty"`
}

type MetricsReporterConfig struct {
	// Report on shutdown (Boolean)
	ReportOnShutdown bool `json:"reportOnShutdown,omitempty"`
	// Output destination - "stdout" or "stderr"
	Output string `json:"output"`
	// Report periodically (Boolean)
	ReportPeriodically bool `json:"reportPeriodically,omitempty"`
	// Interval between consequence reports (in seconds)
	RepotInterval int `json:"reportInterval"`
}

func GetOrDefaultConfig() (*Config, error) {
	return GetOrLoadFromFile("")
}

func GetOrLoadFromFile(path string) (*Config, error) {
	once.Do(func() {
		instance, failure = loadConfig(path)
		return
	})

	return instance, failure
}

func GetOrLoadFromData(data []byte) (*Config, error) {
	once.Do(func() {
		instance, failure = loadFromData(data)
		return
	})

	return instance, failure
}

// Update the defaults when using a configuration structure
func GetOrLoadFromStruct(cfg *Config) (*Config, error) {
	once.Do(func() {
		initDefaults(cfg)
		instance = cfg
		return
	})

	return instance, nil
}

// Eagerly reloads TSDB configuration. Note: not thread-safe
func UpdateConfig(path string) {
	instance, failure = loadConfig(path)
}

// Update the defaults when using an existing configuration structure (custom configuration)
func WithDefaults(cfg *Config) *Config {
	initDefaults(cfg)
	return cfg
}

// Create new configuration structure instance based on given instance.
// All matching attributes within result structure will be overwritten with values of newCfg
func (config *Config) Merge(newCfg *Config) (*Config, error) {
	resultCfg, err := config.merge(newCfg)
	if err != nil {
		return nil, err
	}

	return resultCfg, nil
}

func (config Config) String() string {
	if config.Password != "" {
		config.Password = "SANITIZED"
	}
	if config.AccessKey != "" {
		config.AccessKey = "SANITIZED"
	}

	sanitizedConfigJson, err := json.Marshal(&config)
	if err == nil {
		return string(sanitizedConfigJson)
	} else {
		return fmt.Sprintf("Unable to read config: %v", err)
	}
}

func (*Config) merge(cfg *Config) (*Config, error) {
	mergedCfg := Config{}
	if err := mergo.Merge(&mergedCfg, cfg, mergo.WithOverride); err != nil {
		return nil, errors.Wrap(err, "Unable to merge configurations.")
	}
	return &mergedCfg, nil
}

func loadConfig(path string) (*Config, error) {

	var resolvedPath string

	if strings.TrimSpace(path) != "" {
		resolvedPath = path
	} else {
		envPath := os.Getenv(V3ioConfigEnvironmentVariable)
		if envPath != "" {
			resolvedPath = envPath
		}
	}

	if resolvedPath == "" {
		resolvedPath = DefaultConfigurationFileName
	}

	var data []byte
	if _, err := os.Stat(resolvedPath); err != nil {
		if os.IsNotExist(err) {
			data = []byte{}
		} else {
			return nil, errors.Wrap(err, "Failed to read the TSDB configuration.")
		}
	} else {
		data, err = ioutil.ReadFile(resolvedPath)
		if err != nil {
			return nil, err
		}

		if len(data) == 0 {
			return nil, errors.Errorf("Configuration file '%s' exists but its content is invalid.", resolvedPath)
		}
	}

	return loadFromData(data)
}

func loadFromData(data []byte) (*Config, error) {
	cfg := Config{
		BuildInfo: BuildMetadta,
	}
	err := yaml.Unmarshal(data, &cfg)

	if err != nil {
		return nil, err
	}

	initDefaults(&cfg)

	return &cfg, err
}

func initDefaults(cfg *Config) {
	if cfg.BuildInfo == nil {
		cfg.BuildInfo = BuildMetadta
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = DefaultLogLevel
	}

	// Initialize the default number of workers
	if cfg.ScannerParallelism == 0 {
		cfg.ScannerParallelism = defaultScannerParallelism
	}

	if cfg.DefaultTimeoutInSeconds == 0 {
		cfg.DefaultTimeoutInSeconds = int(defaultTimeoutInSeconds)
	}

	if cfg.WebApiEndpoint == "" {
		cfg.WebApiEndpoint = os.Getenv("V3IO_API")
	}

	if cfg.AccessKey == "" {
		cfg.AccessKey = os.Getenv("V3IO_ACCESS_KEY")
	}

	if cfg.Username == "" {
		cfg.Username = os.Getenv("V3IO_USERNAME")
	}

	if cfg.Password == "" {
		cfg.Password = os.Getenv("V3IO_PASSWORD")
	}
}
