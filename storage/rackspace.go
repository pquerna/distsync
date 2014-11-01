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

package storage

import (
	"github.com/pquerna/distsync/common"
	"github.com/pquerna/distsync/crypto"
	"github.com/rackspace/gophercloud"
	osObjects "github.com/rackspace/gophercloud/openstack/objectstorage/v1/objects"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace"
	"github.com/rackspace/gophercloud/rackspace/objectstorage/v1/objects"

	"errors"
	"io"
	"time"
)

const (
	// This is basically time.RFC3339Nano, but missing the timezone.
	// Swift docs say its always in UTC though. So. Here. We. Go.
	swiftTimelayout = "2006-01-02T15:04:05.999999999"
)

type CloudFilesStorage struct {
	bucket string
	creds  *common.RackspaceCreds
}

func NewCloudFiles(creds *common.RackspaceCreds, bucket string) (*CloudFilesStorage, error) {
	if creds == nil {
		return nil, errors.New("CloudFiles: No Rackspace credentials provided.")
	}

	if bucket == "" {
		return nil, errors.New("CloudFiles: empty StorageBucket")
	}

	return &CloudFilesStorage{
		bucket: bucket,
		creds:  creds,
	}, nil
}

func (cf *CloudFilesStorage) client() (*gophercloud.ServiceClient, error) {
	auth := gophercloud.AuthOptions{
		Username: cf.creds.Username,
		APIKey:   cf.creds.ApiKey,
	}

	ac, err := rackspace.AuthenticatedClient(auth)
	if err != nil {
		return nil, err
	}

	// TOOD: auto-detect serviceNet?
	return rackspace.NewObjectStorageV1(ac, gophercloud.EndpointOpts{
		Region: cf.creds.Region,
	})
}

func (cf *CloudFilesStorage) Download(filename string, writer io.Writer) error {
	client, err := cf.client()
	if err != nil {
		return err
	}

	resp := objects.Download(client, cf.bucket, filename, &osObjects.DownloadOpts{})
	if resp.Err != nil {
		return resp.Err
	}
	defer resp.Body.Close()

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (cf *CloudFilesStorage) List(dc crypto.Decryptor) ([]*FileInfo, error) {
	client, err := cf.client()
	if err != nil {
		return nil, err
	}

	rv := make([]*FileInfo, 0)
	err = objects.List(client, cf.bucket, osObjects.ListOpts{Full: true}).EachPage(func(p pagination.Page) (bool, error) {
		objs, err := objects.ExtractInfo(p)
		if err != nil {
			return false, err
		}
		for _, obj := range objs {
			if obj.Name == ".distsync" {
				continue
			}

			lm, err := time.Parse(swiftTimelayout, obj.LastModified)
			if err != nil {
				return false, err
			}
			name := ""
			if dc != nil {
				name, err = dc.DecryptName(obj.Name)
				if err != nil {
					return false, err
				}
			}

			rv = append(rv, &FileInfo{
				EncryptedName: obj.Name,
				Name:          name,
				LastModified:  lm,
				Length:        int64(obj.Bytes),
			})
		}
		return true, nil
	})

	return rv, nil
}

func (cf *CloudFilesStorage) Upload(filename string, reader io.ReadSeeker) error {
	return errors.New("not done")
}

func (cf *CloudFilesStorage) Start() error {
	return nil
}

func (cf *CloudFilesStorage) Stop() error {
	return nil
}
