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

package commands

import (
	"fmt"
	"net/url"
	"strings"
	"v3io-backup/internal/pkg/performance"
	"v3io-backup/pkg/config"
	"v3io-backup/pkg/utils"

	"github.com/nuclio/logger"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type CmdRoot struct {
	logger      logger.Logger
	cfg         *config.Config
	cmd         *cobra.Command
	v3ioUrl     string
	container   string
	path        string
	cfgFilePath string
	logLevel    string
	username    string
	password    string
	accessKey   string
	Reporter    *performance.MetricReporter
	BuildInfo   *config.BuildInfo
}

func NewCmdRoot() (*CmdRoot, error) {
	commandeer := &CmdRoot{
		BuildInfo: config.BuildMetadta,
	}

	cmd := &cobra.Command{
		Use:          "v3io-backup [command] [arguments] [flags]",
		Short:        "V3IO backup command-line interface (CLI)",
		SilenceUsage: true,
	}

	cmd.PersistentFlags().StringVarP(&commandeer.logLevel, "log-level", "v", "",
		"Verbose output. Add \"=<level>\" to set the log level -\ndebug | info | warn | error. For example: -v=warn.\n(default - \""+config.DefaultLogLevel+"\" when using the flag; \""+config.DefaultLogLevel+"\" otherwise)")
	cmd.PersistentFlags().Lookup("log-level").NoOptDefVal = config.DefaultLogLevel
	cmd.PersistentFlags().StringVarP(&commandeer.path, "path", "d", "",
		"[Required] Path to backup within the configured\ndata container. Examples: \"/my-data\"; \"/tsdb/table-1\".")
	// We don't enforce this flag (commands.MarkFlagRequired("table-path")),
	// although it's documented as Required, because this flag isn't required
	// for the hidden `time` command + during internal tests we might want to
	// configure the table path in a configuration file.
	cmd.PersistentFlags().StringVarP(&commandeer.v3ioUrl, "server", "s", "",
		"Web-gateway (web-APIs) service endpoint of an instance of\nthe Iguazio Continuous Data Platform, of the format\n\"<IP address>:<port number=8081>\". Examples: \"localhost:8081\"\n(when running on the target platform); \"192.168.1.100:8081\".")
	cmd.PersistentFlags().StringVarP(&commandeer.cfgFilePath, "config", "g", "",
		"Path to a YAML backup configuration file. When this flag isn't\nset, the CLI checks for a "+config.DefaultConfigurationFileName+" configuration\nfile in the current directory. CLI flags override file\nconfigurations. Example: \"~/cfg/my_v3io_backup_cfg.yaml\".")
	cmd.PersistentFlags().StringVarP(&commandeer.container, "container", "c", "",
		"The name of an Iguazio Continuous Data Platform data container. Example: \"bigdata\".")
	cmd.PersistentFlags().StringVarP(&commandeer.username, "username", "u", "",
		"Username of an Iguazio Continuous Data Platform user.")
	cmd.PersistentFlags().StringVarP(&commandeer.password, "password", "p", "",
		"Password of the configured user (see -u|--username).")
	cmd.PersistentFlags().StringVarP(&commandeer.accessKey, "access-key", "k", "",
		"Access-key for accessing the required table.\nIf access-key is passed, it will take precedence on user/password authentication.")

	// Add children
	cmd.AddCommand(
		newVersionCmd(commandeer).cmd,
	)

	logger, err := utils.NewLogger(commandeer.logLevel)
	if err != nil {
		return nil, err
	}

	commandeer.logger = logger
	commandeer.cmd = cmd

	return commandeer, nil
}

// Execute the command using os.Args
func (rc *CmdRoot) Execute() error {
	return rc.cmd.Execute()
}

// Return the underlying Cobra command
func (rc *CmdRoot) GetCmd() *cobra.Command {
	return rc.cmd
}

// Generate Markdown files in the target path
func (rc *CmdRoot) CreateMarkdown(path string) error {
	return doc.GenMarkdownTree(rc.cmd, path)
}

func (rc *CmdRoot) initialize() error {
	cfg, err := config.GetOrLoadFromFile(rc.cfgFilePath)
	if err != nil {
		// Display an error if we fail to load a configuration file
		if rc.cfgFilePath == "" {
			return errors.Wrap(err, "Failed to load the TSDB configuration.")
		} else {
			return errors.Wrap(err, fmt.Sprintf("Failed to load the TSDB configuration from '%s'.", rc.cfgFilePath))
		}
	}
	return rc.populateConfig(cfg)
}

func (rc *CmdRoot) populateConfig(cfg *config.Config) error {
	// Initialize performance monitoring
	// TODO: support custom report writers (file, syslog, Prometheus, etc.)
	rc.Reporter = performance.ReporterInstanceFromConfig(cfg)

	if rc.username != "" {
		cfg.Username = rc.username
	}

	if rc.password != "" {
		cfg.Password = rc.password
	}

	if rc.accessKey != "" {
		cfg.AccessKey = rc.accessKey
	}

	if rc.v3ioUrl != "" {
		cfg.WebApiEndpoint = rc.v3ioUrl
	}
	if rc.container != "" {
		cfg.Container = rc.container
	}
	if rc.path != "" {
		cfg.Path = rc.path
	}
	if cfg.WebApiEndpoint == "" {
		return errors.New("web API endpoint must be set")
	}
	if cfg.Container == "" {
		return errors.New("container must be set")
	}
	if cfg.Path == "" {
		return errors.New("table path must be set")
	}
	if rc.logLevel != "" {
		cfg.LogLevel = rc.logLevel
	} else {
		cfg.LogLevel = config.DefaultLogLevel
	}

	// Prefix http:// in case that WebApiEndpoint is a pseudo-URL missing a scheme (for backward compatibility).
	amendedWebApiEndpoint, err := buildUrl(cfg.WebApiEndpoint)
	if err == nil {
		cfg.WebApiEndpoint = amendedWebApiEndpoint
	}

	rc.cfg = cfg
	return nil
}

func buildUrl(webApiEndpoint string) (string, error) {
	if !strings.HasPrefix(webApiEndpoint, "http://") && !strings.HasPrefix(webApiEndpoint, "https://") {
		webApiEndpoint = "http://" + webApiEndpoint
	}
	endpointUrl, err := url.Parse(webApiEndpoint)
	if err != nil {
		return "", err
	}
	endpointUrl.Path = ""
	return endpointUrl.String(), nil
}
