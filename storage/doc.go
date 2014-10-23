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

	"errors"
	"io"
	"strings"
)

type Uploader interface {
	// Upload with this remote filename.
	// See https://code.google.com/p/go/issues/detail?id=6738 for discussion
	// of sized / length'ed readers -- this uses .Seek to calcualte
	// the file size.
	Upload(filename string, reader io.ReadSeeker) error
}

type Downloader interface {
	// Downloads remote filename to io.Writer.
	Download(filename string, reader io.Writer) error
}

// TODO: hash of file? Other attributes?
type FileInfo struct {
	EncryptedName string
	Name          string
	// TODO: LastModified -> non-string
	LastModified string
}

type Lister interface {
	// Returns a list of available files to download.
	List() ([]FileInfo, error)
}

type Storage interface {
	Uploader
	Downloader
	Lister
}

type SizeReader interface {
	io.Reader
	// returns remaining bytes.
	// TODO: should this just be Size() or Len()?
	DistsyncSize() int64
}

func NewFromConf(c *common.Conf) (Storage, error) {
	switch strings.ToUpper(c.Storage) {
	case "S3":
		return NewS3(&c.AwsCreds, c.StorageBucket)
	}

	return nil, errors.New("Unknown storage backend: " + c.Storage)
}
