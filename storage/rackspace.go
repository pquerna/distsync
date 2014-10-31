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

	"errors"
	"io"
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

func (cf *CloudFilesStorage) Download(filename string, writer io.Writer) error {
	return errors.New("not done")
}

func (cf *CloudFilesStorage) List(dc crypto.Decryptor) ([]*FileInfo, error) {
	return nil, errors.New("not done")
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
