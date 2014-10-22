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
	"github.com/mitchellh/go-homedir"
	"github.com/pquerna/distsync/common"
	"github.com/pquerna/distsync/crypto"
	"github.com/pquerna/distsync/storage"

	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Upload struct {
	// bleh, should change this to use channels and shit.
	stop error
	conf *common.Conf
	Ui   cli.Ui
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

type stopError struct{}

func (e *stopError) Error() string {
	return "operation stopped"
}

func (c *Upload) uploadFile(wg *sync.WaitGroup, file string) {
	defer wg.Done()

	err := c._uploadFile(file)
	if err != nil {
		_, ok := err.(*stopError)

		if !ok {
			c.stop = err
		}
	}
}

func (c *Upload) _uploadFile(filename string) error {
	// not proud of this function.
	if c.stop != nil {
		return &stopError{}
	}

	ec, err := crypto.NewFromConf(c.conf)

	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	fpath, err := homedir.Expand(filename)
	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	fpath, err = filepath.Abs(fpath)
	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	_, shortName := filepath.Split(fpath)

	file, err := os.Open(fpath)
	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	tmpFile, err := ioutil.TempFile("", ".distsync")
	if err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if c.stop != nil {
		return &stopError{}
	}

	err = ec.Encrypt(file, tmpFile)
	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	s, err := storage.NewFromConf(c.conf)
	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	// TODO: channel for cancellation of upload?
	err = s.Upload(shortName, tmpFile)
	if err != nil {
		return err
	}

	return nil
}

func (c *Upload) Run(args []string) int {
	var confFile string

	cmdFlags := flag.NewFlagSet("upload", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&confFile, "conf", "~/.distsync", "Configuration path.")

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
		c.Ui.Error("At least one file to upload must be specified.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	var wg sync.WaitGroup

	for _, file := range files {
		c.Ui.Info("Uploading " + file)

		wg.Add(1)
		go c.uploadFile(&wg, file)
	}
	wg.Wait()

	if c.stop != nil {
		c.Ui.Error("Upload failed: " + c.stop.Error())
		c.Ui.Error("")
		return 1
	}
	return 0
}

func (c *Upload) Synopsis() string {
	return "Upload files to distsync"
}
