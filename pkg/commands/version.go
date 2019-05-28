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

	"github.com/spf13/cobra"
)

type cmdVersion struct {
	cmd            *cobra.Command
	rootCommandeer *CmdRoot
}

func newVersionCmd(rc *CmdRoot) *cmdVersion {
	commandeer := &cmdVersion{
		rootCommandeer: rc,
	}

	cmd := &cobra.Command{
		Aliases: []string{"ver"},
		Use:     "version",
		Hidden:  false,
		Short:   "Displays version information",
		Example: "- v3io-backup version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("v3io-backup build details:\n  Build time: %s\n  OS: %s\n  Architecture: %s\n  Version: %s\n  Commit Hash: %s\n  Branch: %s\n",
				rc.BuildInfo.BuildTime,
				rc.BuildInfo.Os,
				rc.BuildInfo.Architecture,
				rc.BuildInfo.Version,
				rc.BuildInfo.CommitHash,
				rc.BuildInfo.Branch)

			return nil
		},
	}

	commandeer.cmd = cmd

	return commandeer
}
