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
	"github.com/pquerna/distsync/crypto"
	"github.com/pquerna/distsync/notify"
	"github.com/pquerna/distsync/storage"

	"flag"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"time"
)

type Daemon struct {
	wg        sync.WaitGroup
	mtx       sync.Mutex
	Ui        cli.Ui
	mainerr   error
	dq        *storage.DownloadQueue
	dl        storage.PersistentDownloader
	notify    notify.Notifier
	conf      *common.Conf
	files     map[string]*storage.FileDownload
	donefiles chan *storage.FileDownload
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

func overwriteFile(name string, t time.Time) bool {
	st, err := os.Stat(name)

	if err != nil {
		if os.IsNotExist(err) {
			return true
		}

		log.WithFields(log.Fields{
			"name": name,
			"st":   st,
			"err":  err,
		}).Error("error stat'ing path")
		return false
	}

	if st.ModTime().After(t) || st.ModTime().Equal(t) {
		log.WithFields(log.Fields{
			"name":         name,
			"local_mtime":  st.ModTime().UTC(),
			"origin_mtime": t.UTC(),
		}).Debug("local file is >= origin file, skipping.")
		return false
	}

	return true
}

func (c *Daemon) updateFiles() error {
	ec, err := crypto.NewFromConf(c.conf)
	if err != nil {
		return err
	}

	st, err := storage.NewFromConf(c.conf)
	if err != nil {
		return err
	}

	workDir, err := homedir.Expand(*c.conf.OutputDir)
	if err != nil {
		return err
	}

	files, err := st.List(ec)

	c.mtx.Lock()
	defer c.mtx.Unlock()

	count := 0

	for _, file := range files {
		fullname := path.Join(workDir, file.Name)

		if overwriteFile(fullname, file.LastModified) == false {
			continue
		}

		fio, ok := c.files[file.Name]
		if ok {
			if file.LastModified.Equal(fio.FileInfo.LastModified) ||
				file.LastModified.Before(fio.FileInfo.LastModified) {
				// same file, but it was older or equal.
				continue
			} else {
				fio.Stop()
			}
		}

		log.WithFields(log.Fields{
			"file": file.Name,
		}).Info("Starting download of file")

		fd := c.dq.Add(c.conf, file, c.donefiles)
		c.files[file.Name] = fd

		count++
	}
	log.WithFields(log.Fields{
		"files_to_download": count,
	}).Info("Completed check of local files.")

	return nil
}

func (c *Daemon) mainLoop() {
	c.files = make(map[string]*storage.FileDownload)
	c.donefiles = make(chan *storage.FileDownload)
	c.dq = storage.NewDownloadQueue(c.dl)

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

	c.mainerr = c.dq.Start()
	if c.mainerr != nil {
		return
	}

	// TODO: fix version number in one place.
	log.WithFields(log.Fields{
		"version":     "0.1.0-dev",
		"working_dir": *c.conf.OutputDir,
	}).Info("distsync daemon started")

	for {
		select {
		case df := <-c.donefiles:
			log.WithFields(log.Fields{
				"file":          df.FileInfo.Name,
				"transfer_rate": df.TransferRate(),
			}).Info("Completed file")
		case <-nchan:
			log.Info("Checking for new files")
			go func() {
				err := c.updateFiles()
				if err != nil {
					c.mainerr = err
					close(nchan)
					return
				}
			}()
		case <-interrupt:
			log.Info("Caught CTRL+C, stopping")
			go func() {
				time.Sleep(500 * time.Millisecond)
				os.Exit(1)
			}()
			return
		}
	}
}

func (c *Daemon) Synopsis() string {
	return "Run's distysnc in Daemon mode to download files."
}
