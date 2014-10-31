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
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
	"github.com/pquerna/distsync/common"
	"github.com/pquerna/distsync/setup"

	"flag"
	_ "fmt"
	"io/ioutil"
	"os"
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

	backend, err := c.pickBackend()
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	var clientconf *common.Conf
	var servconf *common.Conf

	c.Ui.Info("")
	c.Ui.Info("")
	switch backend {
	case BACKEND_AWS:
		clientconf, servconf, err = setup.AWS(c.Ui)
		if err != nil {
			c.Ui.Error("Error: " + err.Error())
			c.Ui.Error("")
			return 1
		}
	case BACKEND_RACKSPACE:
		clientconf, servconf, err = setup.Rackspace(c.Ui)
		if err != nil {
			c.Ui.Error("Error: " + err.Error())
			c.Ui.Error("")
			return 1
		}
	}

	cstr, err := clientconf.ToString()
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	sstr, err := servconf.ToString()
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	c.Ui.Info("")

	err = c.writeConfFile("~/.distsync", cstr)
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	err = c.writeConfFile("~/.distsyncd", sstr)
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	c.Ui.Info("")
	c.Ui.Info("Setup Complete!")
	c.Ui.Info("")
	c.Ui.Info("To get started using distync:")
	c.Ui.Info("1) Copy ~/.distsyncd to your servers, and run:")
	c.Ui.Info("  distsync daemon")
	c.Ui.Info("")
	c.Ui.Info("")
	c.Ui.Info("2) Use `distsync upload` from your build infrastructure:")
	c.Ui.Info("")
	c.Ui.Info("  distsync upload myapp-1.0.tar.gz")
	c.Ui.Info("")
	c.Ui.Info("")

	return 0
}

type backend int

const (
	BACKEND_NONE      backend = 0
	BACKEND_AWS               = 1
	BACKEND_RACKSPACE         = 2
)

func (c *Setup) pickBackend() (backend, error) {

	// TODO: more choices, use array offset
	be, err := common.Choice(c.Ui, "What Cloud Service should distsync use to store files?", []string{
		"AWS S3",
		"Rackspace Cloud Files",
	})

	if err != nil {
		return BACKEND_NONE, err
	}
	return backend(be + 1), nil
}

func fileexists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (c *Setup) writeConfFile(name string, contents string) error {
	name, err := homedir.Expand(name)
	if err != nil {
		return err
	}

	if fileexists(name) {
		yn, err := common.YesNoChoice(c.Ui,
			"File "+name+" exists.\nOverwrite? [y/n]")
		if err != nil {
			return err
		}

		if yn == false {
			name, err = c.Ui.Ask("Alternative filename? ")
			if err != nil {
				return err
			}
			return c.writeConfFile(name, contents)
		}
	}

	err = ioutil.WriteFile(name, []byte(contents), 0600)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"path": name,
	}).Info("Wrote config file.")
	return nil
}

func (c *Setup) Synopsis() string {
	return "Configures distsync and creates required cloud resources."
}
