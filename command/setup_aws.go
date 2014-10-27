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
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/iam"
	"github.com/mitchellh/goamz/s3"
	"github.com/pquerna/distsync/common"
	"github.com/pquerna/distsync/crypto"

	"encoding/json"
	"sort"
)

func awsCreateUser(client *iam.IAM, name string, policy string) (*iam.AccessKey, error) {
	log.WithFields(log.Fields{
		"username": name,
	}).Info("Creating User")

	_, err := client.CreateUser(name, "/")
	//	user := userResp.User
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"username": name,
	}).Info("Applying restricted IAM Policy")

	_, err = client.PutUserPolicy(name, "distsync-policy", policy)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"username": name,
	}).Info("Creating Access Key")

	ak, err := client.CreateAccessKey(name)
	if err != nil {
		return nil, err
	}

	return &ak.AccessKey, nil
}

func (c *Setup) setupAws() (*common.Conf, *common.Conf, error) {
	dsId := uuid.NewRandom().String()
	bucketName := "distsync-" + dsId

	sharedSecret, err := crypto.RandomSecret()
	if err != nil {
		return nil, nil, err
	}

	auth, err := c.awsAuth()
	if err != nil {
		return nil, nil, err
	}

	region, err := c.awsRegion()
	if err != nil {
		return nil, nil, err
	}

	s3Client := s3.New(auth, region)

	bucket := s3Client.Bucket(bucketName)

	log.WithFields(log.Fields{
		"bucket": bucketName,
		"region": region.Name,
	}).Info("Creating S3 Bucket")

	err = bucket.PutBucket("private")
	if err != nil {
		return nil, nil, err
	}

	iamClient := iam.New(auth, region)

	uploader := "distsync-upload-" + dsId
	downloader := "distsync-download-" + dsId

	policyUploader, err := policyUploader(bucketName)
	if err != nil {
		return nil, nil, err
	}

	policyDownloader, err := policyDownloader(bucketName)
	if err != nil {
		return nil, nil, err
	}

	akUp, err := awsCreateUser(iamClient, uploader, policyUploader)
	if err != nil {
		return nil, nil, err
	}
	log.WithFields(log.Fields{
		"username":  uploader,
		"accesskey": akUp.Id,
	}).Info("Created User and AccessKey for uploading")

	clientconf := common.NewConf()
	clientconf.SharedSecret = sharedSecret
	clientconf.StorageBucket = bucketName
	clientconf.Aws = &common.AwsCreds{
		Region:    region.Name,
		AccessKey: akUp.Id,
		SecretKey: akUp.Secret,
	}

	akDown, err := awsCreateUser(iamClient, downloader, policyDownloader)
	if err != nil {
		return nil, nil, err
	}
	log.WithFields(log.Fields{
		"username":  downloader,
		"accesskey": akDown.Id,
	}).Info("Created User and AccessKey for downloading")

	serverconf := common.NewConf()
	serverconf.SharedSecret = sharedSecret
	serverconf.StorageBucket = bucketName
	outdir := "~/"
	serverconf.OutputDir = &outdir
	serverconf.Aws = &common.AwsCreds{
		Region:    region.Name,
		AccessKey: akDown.Id,
		SecretKey: akDown.Secret,
	}

	/*
		conf.PeerDist = &common.PeerDist{
			ListenAddr: ":4166",
			GossipAddr: ":4166",
		}
	*/

	return clientconf, serverconf, nil
}

func (c *Setup) awsRegion() (aws.Region, error) {
	regions := make(sort.StringSlice, 0)

	for k, _ := range aws.Regions {
		regions = append(regions, k)
	}

	regions.Sort()

	choice, err := common.Choice(c.Ui, "Which AWS Region should distsync upload files to?", regions)
	if err != nil {
		return aws.Region{}, err
	}

	return aws.Regions[regions[choice]], nil
}

func (c *Setup) awsAuth() (aws.Auth, error) {
	auth, err := aws.SharedAuth()
	if err == nil {
		return auth, nil
	}

	auth, err = aws.EnvAuth()
	if err == nil {
		return auth, nil
	}

	auth, err = c.awsPromptAuth()
	if err == nil {
		return auth, nil
	}
	return aws.Auth{}, err

}

func (c *Setup) awsPromptAuth() (aws.Auth, error) {
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

func policyBuilder(actions []string, resources []string) (string, error) {
	p := IAMPolicy{
		Version: "2012-10-17",
		Statement: []IAMStatement{
			IAMStatement{
				Effect:   "Allow",
				Action:   actions,
				Resource: resources,
			},
		},
	}

	b, err := json.Marshal(p)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func policyUploader(bucket string) (string, error) {
	return policyBuilder(
		[]string{
			"s3:ListBucket",
			"s3:PutObject",
		},
		[]string{
			"arn:aws:s3:::" + bucket + "",
			"arn:aws:s3:::" + bucket + "/*",
		})
}

func policyDownloader(bucket string) (string, error) {
	return policyBuilder(
		[]string{
			"s3:ListBucket",
			"s3:GetObject",
			"s3:GetObjectTorrent",
		},
		[]string{
			"arn:aws:s3:::" + bucket + "",
			"arn:aws:s3:::" + bucket + "/*",
		})
}
