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

	"io"
)

// A notifier provides a channel for when
// the Manifest has changed.
type Uploader interface {
	// Upload with this remote filename.
	Upload(filename string, reader io.Reader) error
}

type Downloader interface {
	// Downloads remote filename to io.Writer.
	Download(filename string, reader io.Writer) error
}

// TODO: hash of file? Other attributes?
type FileInfo struct {
	Name string
}

type Lister interface {
	// Returns a list of available files to download.
	List() []FileInfo
}

// Uploads to S3, and touches .distsync on success,
// which `notify.S3Poller` uses to find changes.
//func NewS3Uploader(common.S3Conf) Uploader {
//	return nil
//}

func NewUploaderFromConf(c *common.Conf) (Uploader, error) {
	return nil, nil
}
