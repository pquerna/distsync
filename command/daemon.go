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
	"github.com/pquerna/distsync/common"
	"github.com/pquerna/distsync/notify"
	"github.com/pquerna/distsync/storage"

	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
)

type Daemon struct {
	Ui      cli.Ui
	wg      sync.WaitGroup
	mainerr error
	dl      storage.PersistentDownloader
	notify  notify.Notifier
	conf    *common.Conf
}

func (c *Daemon) Help() string {
	helpText := `
Usage: distsync daemon [options]

  Runs distsync in daemon mode.  This will listen for new
  files, and automatically download them to the configured
  path.

Options:

  -conf=~/.distsyncd         Read specific configuration file.
`
	return strings.TrimSpace(helpText)
}

func (c *Daemon) Run(args []string) int {
	var confFile string

	cmdFlags := flag.NewFlagSet("daemon", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&confFile, "conf", "~/.distsyncd", "Configuration path.")

	err := cmdFlags.Parse(args)
	if err != nil {
		return 1
	}

	c.conf, err = common.ConfFromFile(confFile)
	if err != nil {
		c.Ui.Error("Configuration failure: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	if len(cmdFlags.Args()) != 0 {
		c.Ui.Error("daemon takes no arguments.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	c.dl, err = storage.NewPersistentDownloader(c.conf)
	if err != nil {
		c.Ui.Error("Error configuring downloader: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	c.notify, err = notify.NewFromConf(c.conf)
	if err != nil {
		c.Ui.Error("Error configuring notify backend: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	c.wg.Add(1)
	go c.mainLoop()
	c.wg.Wait()

	if c.mainerr != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	return 0
}

func (c *Daemon) stop() {
	defer c.wg.Done()
	c.notify.Stop()
	c.dl.Stop()
}

func (c *Daemon) mainLoop() {
	defer c.stop()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c.mainerr = c.dl.Start()
	if c.mainerr != nil {
		return
	}

	nchan := c.notify.Changed()
	c.mainerr = c.notify.Start()
	if c.mainerr != nil {
		return
	}

	for {
		select {
		case <-nchan:
			println("got notification")
		case <-interrupt:
			log.Info("CTL+C, stopping.")
			return
		}
	}
}

func (c *Daemon) Synopsis() string {
	return "Run's distysnc in Daemon mode to download files."
}
