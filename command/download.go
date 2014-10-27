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
	"github.com/pquerna/distsync/crypto"
	"github.com/pquerna/distsync/storage"

	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
)

type Download struct {
	Ui        cli.Ui
	stop      error
	dl        storage.PersistentDownloader
	dq        *storage.DownloadQueue
	donefiles chan *storage.FileDownload
	conf      *common.Conf
}

func (c *Download) Help() string {
	helpText := `
Usage: distsync download [options] file

  Downloads specified files to your local directory.

Options:

  -conf=~/.distsyncd         Read specific configuration file.
`
	return strings.TrimSpace(helpText)
}

func diffslice(a []string, b []string) []string {
	rv := make([]string, 0)
	tmp := make(map[string]int)

	for _, v := range a {
		tmp[v] = 1
	}

	for _, v := range b {
		tmp[v] += 1
	}

	for k, v := range tmp {
		if v == 1 {
			rv = append(rv, k)
		}
	}

	return rv
}

func (c *Download) Run(args []string) int {
	var confFile string

	cmdFlags := flag.NewFlagSet("download", flag.ContinueOnError)
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

	files := cmdFlags.Args()
	if len(files) == 0 {
		c.Ui.Error("At least one file to download must be specified.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	download, err := c.getFilesToDownload(files)
	if err != nil {
		c.Ui.Error("Error Getting files to download: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	if len(download) != len(files) {
		dlfiles := make([]string, 0, len(download))
		for _, dl := range download {
			dlfiles = append(dlfiles, dl.Name)
		}
		diff := diffslice(files, dlfiles)
		c.Ui.Error(fmt.Sprintf("Files not found: %v", diff))
		c.Ui.Error("")
		return 1
	}

	cwd, err := os.Getwd()
	if err != nil {
		c.Ui.Error("Error getting cwd: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	c.dl, err = storage.NewPersistentDownloader(c.conf)
	if err != nil {
		c.Ui.Error("Error configuring download: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	err = c.dl.Start()
	if err != nil {
		c.Ui.Error("Error starting download: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	c.dq = storage.NewDownloadQueue(c.dl)

	err = c.dq.Start()
	if err != nil {
		c.Ui.Error("Error starting download queue: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	c.donefiles = make(chan *storage.FileDownload)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c.conf.OutputDir = &cwd

	for _, v := range download {
		c.dq.Add(c.conf, v, c.donefiles)
	}

	completed := 0
	defer func() {
		c.dl.Stop()
		c.dq.Stop()
	}()

	for {
		select {
		case done := <-c.donefiles:
			completed += 1
			if done.Error != nil {
				c.Ui.Error("Error downloading " + done.FileInfo.Name + " :" + err.Error())
				c.Ui.Error("")
			} else {
				c.Ui.Info("Download Complete: " + done.FileInfo.Name)
			}

			if completed == len(download) {
				return 0
			}
		case <-interrupt:
			c.Ui.Info("Caught CTRL+C, stopping....")
			return 1
		}
	}

	if c.stop != nil {
		c.Ui.Error("Download failed: " + c.stop.Error())
		c.Ui.Error("")
		return 1
	}
	return 0
}

func (c *Download) getFilesToDownload(fnames []string) ([]*storage.FileInfo, error) {
	ec, err := crypto.NewFromConf(c.conf)
	if err != nil {
		return nil, err
	}

	s, err := storage.NewFromConf(c.conf)
	if err != nil {
		return nil, err
	}

	storedFiles, err := s.List(ec)
	if err != nil {
		return nil, err
	}

	download := make([]*storage.FileInfo, 0, 1)
	for _, file := range storedFiles {
		for _, fname := range fnames {
			// TODO: meh.
			if file.Name == fname {
				download = append(download, file)
			}
		}
	}

	return download, nil
}

func (c *Download) Synopsis() string {
	return "Downloads files from distsync"
}
