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
	"github.com/pquerna/distsync/encrypt"

	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type TestCrypt struct {
	Ui cli.Ui
}

func (c *TestCrypt) Help() string {
	helpText := `
Usage: distsync test-crypt [options] file

  Performs crypto operation on a file.

Options:

  -e	Encrypt
  -d	Decrypt
`
	return strings.TrimSpace(helpText)
}

func (c *TestCrypt) Run(args []string) int {
	var encryptOpt bool
	var decryptOpt bool

	cmdFlags := flag.NewFlagSet("test-crypt", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.BoolVar(&encryptOpt, "e", false, "should encrypt")
	cmdFlags.BoolVar(&decryptOpt, "d", false, "should encrypt")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	files := cmdFlags.Args()
	if len(files) != 1 {
		c.Ui.Error("Only one file can be specified.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	if encryptOpt == decryptOpt {
		c.Ui.Error("Pick one, encrypt or decrypt.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	fpath, err := homedir.Expand(files[0])
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	fpath, err = filepath.Abs(fpath)
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	file, err := os.Open(fpath)
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	dname := filepath.Dir(fpath)

	ec, err := encrypt.NewEtmCryptor([]byte("hellohelloworld1hellohelloworld1"))
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	if encryptOpt {
		destpath := fpath + ".denc"
		tmpFile, err := ioutil.TempFile(dname, ".distsync")
		if err != nil {
			c.Ui.Error("Error: " + err.Error())
			c.Ui.Error("")
			return 1
		}

		tmpName := tmpFile.Name()

		err = ec.Encrypt(file, tmpFile)
		if err != nil {
			c.Ui.Error("Error: " + err.Error())
			c.Ui.Error("")
			return 1
		}

		err = tmpFile.Close()
		if err != nil {
			c.Ui.Error("Error: " + err.Error())
			c.Ui.Error("")
			return 1
		}

		os.Rename(tmpName, destpath)
	}

	if decryptOpt {
		destpath := strings.TrimSuffix(fpath, ".denc")
		if destpath == fpath {
			c.Ui.Error("Error: Source file must end in .denc")
			c.Ui.Error("")
			return 1
		}

		tmpFile, err := ioutil.TempFile(dname, ".distsync")
		if err != nil {
			c.Ui.Error("Error: " + err.Error())
			c.Ui.Error("")
			return 1
		}

		tmpName := tmpFile.Name()

		err = ec.Decrypt(file, tmpFile)
		if err != nil {
			c.Ui.Error("Error: " + err.Error())
			c.Ui.Error("")
			return 1
		}

		err = tmpFile.Close()
		if err != nil {
			c.Ui.Error("Error: " + err.Error())
			c.Ui.Error("")
			return 1
		}

		os.Rename(tmpName, destpath)
	}

	return 0
}

func (c *TestCrypt) Synopsis() string {
	return "Test distsync's file encryption format."
}
