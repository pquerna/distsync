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

package notify

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/pquerna/distsync/common"

	"errors"
)

type s3Poll struct {
	bucket   string
	lastEtag string
	creds    *common.AwsCreds
}

// Polls the specified S3 bucket for a new .distsync file every 10 to 20 seconds.
//
// $0.004 per 10,000 requests.
// 2.62974e6 seconds in a month.
// 175,316, 15 second periods.
// 17.5316, 10,000 requests bundles.
// 17.5316 * $0.0044 = $0.077 per month per watcher for request charges.
//
func NewS3Poll(conf *common.AwsCreds, bucketName string) (Notifier, error) {
	return newTimedPoller(
		&s3Poll{
			bucket: bucketName,
			creds:  conf,
		}), nil
}

func (s *s3Poll) client() (*s3.S3, error) {
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

func (sp *s3Poll) Poll() (bool, error) {
	client, err := sp.client()
	if err != nil {
		return false, err
	}

	bucket := client.Bucket(sp.bucket)

	log.WithFields(log.Fields{
		"last_etag": sp.lastEtag,
		"bucket":    sp.bucket,
		"file":      ".distsync",
	}).Debug("Checking for changed ETag")

	resp, err := bucket.Head(".distsync")

	if err != nil {
		return false, err
	}

	etag := resp.Header.Get("ETag")
	if etag == "" {
		return false, errors.New("Empty ETag on HEAD")
	}

	if etag != sp.lastEtag {
		log.WithFields(log.Fields{
			"last_etag": sp.lastEtag,
			"new_etag":  etag,
		}).Info("ETag changed, notifying watchers.")
		sp.lastEtag = etag
		return true, nil
	}

	return false, nil
}
