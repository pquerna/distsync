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
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/pquerna/distsync/common"

	"errors"
	"io"
)

type S3Storage struct {
	bucket string
	creds  *common.AwsCreds
}

func NewS3(creds *common.AwsCreds, bucket string) (*S3Storage, error) {
	if bucket == "" {
		return nil, errors.New("S3: empty StorageBucket")
	}

	return &S3Storage{
		bucket: bucket,
		creds:  creds,
	}, nil
}

func (s *S3Storage) client() (*s3.S3, error) {
	a := aws.Auth{
		AccessKey: s.creds.AccessKey,
		SecretKey: s.creds.SecretKey,
	}
	r, ok := aws.Regions[s.creds.Region]
	if !ok {
		return nil, errors.New("S3: Unkonwn region: '" + s.creds.Region + "'")
	}

	return s3.New(a, r), nil
}

var dsyncCt = "application/distsync-encrypted"

func (s *S3Storage) Upload(filename string, reader io.ReadSeeker) error {
	l, err := reader.Seek(2, 0)

	if err != nil {
		return err
	}

	_, err = reader.Seek(0, 0)

	if err != nil {
		return err
	}

	client, err := s.client()
	if err != nil {
		return err
	}

	bucket := client.Bucket(s.bucket)

	err = bucket.PutReader(filename, reader, l, dsyncCt, "")
	if err != nil {
		return err
	}

	return nil
}

func (s *S3Storage) Download(filename string, reader io.Writer) error {
	return nil
}

func (s *S3Storage) List() ([]FileInfo, error) {
	return nil, nil
}
