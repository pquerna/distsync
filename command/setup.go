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
	"code.google.com/p/go-uuid/uuid"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/iam"
	"github.com/mitchellh/goamz/s3"
	"github.com/pquerna/distsync/common"

	"encoding/json"
	"flag"
	_ "fmt"
	"sort"
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

func (c *Setup) getRegion() (aws.Region, error) {
	regions := make(sort.StringSlice, 0)

	for k, _ := range aws.Regions {
		regions = append(regions, k)
	}

	regions.Sort()

	choice, err := common.Choice(c.Ui, "AWS Region?", regions)
	if err != nil {
		return aws.Region{}, err
	}

	return aws.Regions[regions[choice]], nil
}

func (c *Setup) getAuth() (aws.Auth, error) {
	auth, err := aws.SharedAuth()
	if err == nil {
		return auth, nil
	}

	auth, err = aws.EnvAuth()
	if err == nil {
		return auth, nil
	}

	auth, err = c.promptAuth()
	if err == nil {
		return auth, nil
	}
	return aws.Auth{}, err

}

func (c *Setup) promptAuth() (aws.Auth, error) {
	var err error
	a := aws.Auth{}
	a.AccessKey, err = c.Ui.Ask("AWS AccessKey: ")
	if err != nil {
		return a, err
	}

	a.SecretKey, err = c.Ui.Ask("AWS SecretKey: ")
	if err != nil {
		return a, err
	}

	a.Token, err = c.Ui.Ask("AWS Token (optional, press enter for none): ")
	if err != nil {
		return a, err
	}

	return a, nil
}

type s3PolicyInfo struct {
	Name string
}

type IAMStatement struct {
	Effect   string
	Action   []string
	Resource []string
}

type IAMPolicy struct {
	Version   string
	Statement []IAMStatement
}

func s3policy(name string) (string, error) {
	p := IAMPolicy{
		Version: "2012-10-17",
		Statement: []IAMStatement{
			IAMStatement{
				Effect: "Allow",
				Action: []string{
					"s3:*",
				},
				Resource: []string{
					"arn:aws:s3:::\"" + name + "\"",
					"arn:aws:s3:::\"" + name + "/*\"",
				},
			},
		},
	}

	b, err := json.Marshal(p)

	if err != nil {
		return "", err
	}

	return string(b), nil
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

	auth, err := c.getAuth()
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	region, err := c.getRegion()
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	name := "distsync-" + uuid.NewRandom().String()

	s3Client := s3.New(auth, region)

	bucket := s3Client.Bucket(name)
	err = bucket.PutBucket("public-read")
	if err != nil {
		c.Ui.Error("S3 error on bucket creation: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	iamClient := iam.New(auth, region)
	_, err = iamClient.CreateUser(name, "/")
	//	user := userResp.User
	if err != nil {
		c.Ui.Error("IAM Error on CreateUser: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	policy, err := s3policy(name)
	if err != nil {
		c.Ui.Error("Policy Template error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	_, err = iamClient.PutUserPolicy(name, "distsync-uploader", policy)
	if err != nil {
		c.Ui.Error("IAM Error on PutUserPolicy: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	ak, err := iamClient.CreateAccessKey(name)
	if err != nil {
		c.Ui.Error("IAM Error on CreateAccessKey: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	println(ak.AccessKey.Id)
	println(ak.AccessKey.Secret)

	//
	// encryption backend
	// storage backend
	// notification backend
	_, err = common.Choice(c.Ui, "Fav cat?", []string{"garfield", "bob", "Moo"})
	if err != nil {
		c.Ui.Error("Error: " + err.Error())
		c.Ui.Error("")
		return 1
	}

	return 0
}

func (c *Setup) Synopsis() string {
	return "Configure a new distsync."
}
