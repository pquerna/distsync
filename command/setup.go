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
	"github.com/pquerna/distsync/common"

	"flag"
	"strings"
)

type Setup struct {
	Ui cli.Ui
}

func (c *Setup) Help() string {
	helpText := `
Usage: distsync setup

  Prompts you with interactive questions about configuring
  distsync.

`
	return strings.TrimSpace(helpText)
}

func (c *Setup) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("setup", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if len(cmdFlags.Args()) != 0 {
		c.Ui.Error("setup takes no arguments.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	_, err := common.Choice(c.Ui, "Fav cat?", []string{"garfield", "bob", "Moo"})
	if err != nil {
		c.Ui.Error("Setup failed")
		c.Ui.Error("")
		panic(err)
		return 1
	}
	return 0
}

func (c *Setup) Synopsis() string {
	return "Configure a new distsync."
}
