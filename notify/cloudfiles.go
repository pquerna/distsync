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
	"github.com/pquerna/distsync/common"
	"github.com/rackspace/gophercloud"
	osObjects "github.com/rackspace/gophercloud/openstack/objectstorage/v1/objects"
	"github.com/rackspace/gophercloud/rackspace"
	"github.com/rackspace/gophercloud/rackspace/objectstorage/v1/objects"

	"errors"
)

type cloudFilesPoll struct {
	bucket   string
	lastEtag string
	creds    *common.RackspaceCreds
}

func NewCloudFilesPoll(conf *common.RackspaceCreds, bucketName string) (Notifier, error) {
	return newTimedPoller(
		&cloudFilesPoll{
			bucket: bucketName,
			creds:  conf,
		}), nil
}

func (cf *cloudFilesPoll) client() (*gophercloud.ServiceClient, error) {
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

func (cf *cloudFilesPoll) Poll() (bool, error) {
	client, err := cf.client()
	if err != nil {
		return false, err
	}

	log.WithFields(log.Fields{
		"last_etag": cf.lastEtag,
		"bucket":    cf.bucket,
		"file":      ".distsync",
	}).Debug("Checking for changed ETag")

	resp := objects.Download(client, cf.bucket, ".distsync",
		&osObjects.DownloadOpts{
			IfNoneMatch: cf.lastEtag,
		})
	if resp.Err != nil {
		return false, resp.Err
	}
	defer resp.Body.Close()

	headers, err := resp.ExtractHeader()
	if err != nil {
		return false, err
	}

	etag := headers.Get("ETag")
	if etag == "" {
		return false, errors.New("Empty ETag on Request")
	}

	if etag != cf.lastEtag {
		log.WithFields(log.Fields{
			"last_etag": cf.lastEtag,
			"new_etag":  etag,
		}).Info("ETag changed, notifying watchers.")

		cf.lastEtag = etag
		return true, nil
	}

	return false, nil
}
