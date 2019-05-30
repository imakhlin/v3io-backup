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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"v3io-backup/pkg/backend/v3io"
	"v3io-backup/pkg/config"
)

type cmdBackup struct {
	cmd            *cobra.Command
	rootCommandeer *CmdRoot
	paths          []string // comma separated  list of paths to backup the data from in the source container
	excludeFilters []string // comma separated list of filter expressions to be applied on the file names in given path(s)
	targetRepo     string   // The destination repository URL
}

func newBackupCmd(rootCommandeer *CmdRoot) *cmdBackup {
	commandeer := &cmdBackup{
		rootCommandeer: rootCommandeer,
	}

	cmd := &cobra.Command{
		Aliases: []string{"bk"},
		Use:     "backup <target repository URL> <source URL> [<paths>] [<filters>] [flags]",
		Short:   "Backup data from the source to the target repository",
		Long:    `Backup data from given data source onto the target backup repository`,
		Example: `The examples assume that the endpoint of the web-gateway service, the login credentials, and
the name of the data container are configured in the default configuration file (` + config.DefaultConfigurationFileName + `)
instead of using the -s|--server, -u|--username, -p|--password, and -c|--container flags.
- v3io-backup backup ...TBD...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return commandeer.backup()
		},
	}

	cmd.Flags().StringArrayVarP(&commandeer.paths, "paths", "d", []string{"/"},
		"Paths to backup within the configured\ndata container. Examples: \"/my-data\"; \"/home/user/table-1\".")
	cmd.Flags().StringArrayVarP(&commandeer.excludeFilters, "excludes", "e", nil,
		"Comma separated list of filter expressions (RegEx). All matching items will be excluded. Empty by default.")
	cmd.Flags().StringVarP(&commandeer.targetRepo, "repo", "r", "",
		"The target backup repository URL")

	commandeer.cmd = cmd

	return commandeer
}

func (bc *cmdBackup) backup() error {

	if bc.targetRepo == "" {
		return errors.New("The backup command must receive target repository parameters (set via the -r|--repo flag).")
	}

	// Initialize parameters and adapter
	if err := bc.rootCommandeer.initialize(); err != nil {
		return err
	}

	if bc.paths != nil {
		bc.rootCommandeer.cfg.BackupOptions.Paths = bc.paths
	}

	if bc.excludeFilters != nil {
		bc.rootCommandeer.cfg.BackupOptions.ExcludeFilters = bc.excludeFilters
	}

	if bc.excludeFilters != nil {
		bc.rootCommandeer.cfg.BackupOptions.Repository = bc.targetRepo
	}

	logger := bc.rootCommandeer.logger
	logger.InfoWith("Backup", "source", bc.rootCommandeer.v3ioUrl, "paths", bc.paths, "filter", bc.excludeFilters,
		"target repository", bc.targetRepo, "username", bc.rootCommandeer.username, "access-key", bc.rootCommandeer.accessKey, "log-level", bc.rootCommandeer.logLevel)

	ds, err := v3io.NewDataSource(bc.rootCommandeer.cfg)

	if err != nil {
		return err
	}

	err = ds.Connect()
	if err != nil {
		return err
	}

	return nil
}
