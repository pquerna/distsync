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
	"github.com/pquerna/distsync/crypto"

	"errors"
	"io"
	"strings"
	"time"
)

type S3Storage struct {
	bucket string
	creds  *common.AwsCreds
}

func NewS3(creds *common.AwsCreds, bucket string) (*S3Storage, error) {
	if creds == nil {
		return nil, errors.New("S3: No AwsCreds provided.")
	}

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

// Uploads to S3, and touches .distsync on success.
// which `notify.S3Poller` uses to find changes.
func (s *S3Storage) Upload(filename string, reader io.ReadSeeker) error {
	// just a random string taht will change the etag of .distsync,
	// so that `notify.S3Poller` look for new files.
	tsec, err := crypto.RandomSecret()
	if err != nil {
		return err
	}

	l, err := reader.Seek(0, 2)

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

	sr := strings.NewReader(tsec)
	err = bucket.PutReader(".distsync", sr, int64(sr.Len()), "text/plain", "")
	if err != nil {
		return err
	}

	return nil
}

func (s *S3Storage) DownloadTorrent(filename string, writer io.Writer) error {
	client, err := s.client()
	if err != nil {
		return err
	}

	bucket := client.Bucket(s.bucket)

	r, err := bucket.GetTorrentReader(filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, r)
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Storage) Download(filename string, writer io.Writer) error {
	client, err := s.client()
	if err != nil {
		return err
	}

	bucket := client.Bucket(s.bucket)

	r, err := bucket.GetReader(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = io.Copy(writer, r)
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Storage) List(dc crypto.Decryptor) ([]*FileInfo, error) {
	client, err := s.client()
	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(s.bucket)

	contents, err := bucket.GetBucketContents()
	if err != nil {
		return nil, err
	}

	rv := make([]*FileInfo, 0, len(*contents))
	for _, key := range *contents {
		if key.Key == ".distsync" {
			continue
		}

		lm, err := time.Parse(time.RFC3339Nano, key.LastModified)
		if err != nil {
			return nil, err
		}

		name := ""
		if dc != nil {
			name, err = dc.DecryptName(key.Key)
			if err != nil {
				return nil, err
			}
		}

		rv = append(rv, &FileInfo{
			EncryptedName: key.Key,
			Name:          name,
			LastModified:  lm,
		})
	}

	return rv, nil
}

func (s *S3Storage) Start() error {
	return nil
}

func (s *S3Storage) Stop() error {
	return nil
}
