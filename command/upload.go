/**
 *  Copyright 2014 Paul Querna
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package command

import (
	"github.com/mitchellh/cli"

	"flag"
	"strings"
)

type Upload struct {
	Ui cli.Ui
}

func (c *Upload) Help() string {
	helpText := `
Usage: distsync upload [options] file ...

  Uploads specified files to the configured storage area.

Options:

  -conf=~/.distsync         Read specific configuration file.
`
	return strings.TrimSpace(helpText)
}

func (c *Upload) Run(args []string) int {
	var confFile string

	cmdFlags := flag.NewFlagSet("upload", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&confFile, "conf", "~/.distsync", "Configuration path.")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	files := cmdFlags.Args()
	if len(files) == 0 {
		c.Ui.Error("At least one file to upload must be specified.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	return 0
}

func (c *Upload) Synopsis() string {
	return "Upload files to distsync"
}
