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
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type Download struct {
	Ui   cli.Ui
	stop error
	conf *common.Conf
}

func (c *Download) Help() string {
	helpText := `
Usage: distsync download [options] file

  Downloads specified files to your local directory.

Options:

  -conf=~/.distsync         Read specific configuration file.
`
	return strings.TrimSpace(helpText)
}

func (c *Download) Run(args []string) int {
	var confFile string

	cmdFlags := flag.NewFlagSet("download", flag.ContinueOnError)
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
		c.Ui.Error("At least one file to download must be specified.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	download, err := c.getFilesToDownload(files)
	if err != nil {
		c.Ui.Error("Getting files to download failed: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	var wg sync.WaitGroup

	for _, file := range download {
		wg.Add(1)
		go c.downloadFile(&wg, file)
	}
	wg.Wait()

	if c.stop != nil {
		c.Ui.Error("Download failed: " + c.stop.Error())
		c.Ui.Error("")
		return 1
	}
	return 0
}

func (c *Download) getFilesToDownload(fnames []string) ([]storage.FileInfo, error) {
	ec, err := crypto.NewFromConf(c.conf)

	if err != nil {
		return nil, err
	}

	s, err := storage.NewFromConf(c.conf)
	if err != nil {
		return nil, err
	}

	storedFiles, err := s.List()
	if err != nil {
		return nil, err
	}

	download := make([]storage.FileInfo, 0, 1)
	for _, file := range storedFiles {
		if file.EncryptedName == ".distsync" {
			continue
		}

		file.Name, err = ec.DecryptName(file.EncryptedName)
		if err != nil {
			return nil, err
		}

		// TODO: meh
		for _, fname := range fnames {
			if file.Name == fname {
				download = append(download, file)
			}
		}
	}

	return download, nil
}

func (c *Download) downloadFile(wg *sync.WaitGroup, fi storage.FileInfo) {
	defer wg.Done()

	err := c._downloadFile(fi)
	if err != nil {
		_, ok := err.(*stopError)

		if !ok {
			c.stop = err
		}
	}
}

func (c *Download) _downloadFile(fi storage.FileInfo) error {
	ec, err := crypto.NewFromConf(c.conf)

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

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	// TODO: consider io.Pipe() ?
	tmpFileEnc, err := ioutil.TempFile(cwd, ".distsync-e")
	if err != nil {
		return err
	}
	defer func() {
		tmpFileEnc.Close()
		os.Remove(tmpFileEnc.Name())
	}()
	if c.stop != nil {
		return &stopError{}
	}

	c.Ui.Info("Downloading " + fi.EncryptedName + " -> " + fi.Name)

	err = s.Download(fi.EncryptedName, tmpFileEnc)
	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	tmpFile, err := ioutil.TempFile(cwd, ".distsync")
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

	_, err = tmpFileEnc.Seek(0, 0)
	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	err = ec.Decrypt(tmpFileEnc, tmpFile)

	if err != nil {
		return err
	}
	if c.stop != nil {
		return &stopError{}
	}

	err = os.Rename(tmpFile.Name(), fi.Name)
	if err != nil {
		return err
	}

	return nil
}

func (c *Download) Synopsis() string {
	return "Downloads files from distsync"
}
