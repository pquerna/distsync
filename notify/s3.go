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
	"math/rand"
	"os"
	"sync"
	"time"
)

type s3Poll struct {
	mtx      sync.Mutex
	wg       sync.WaitGroup
	bucket   string
	lastEtag string
	creds    *common.AwsCreds
	changes  []chan int
	quit     chan int
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
	return &s3Poll{
		bucket:  bucketName,
		creds:   conf,
		changes: make([]chan int, 0),
		quit:    make(chan int),
	}, nil
}

func (sp *s3Poll) broadcast() {
	sp.mtx.Lock()
	defer sp.mtx.Unlock()

	for _, v := range sp.changes {
		v <- 1
	}
}

func (sp *s3Poll) Changed() chan int {
	c := make(chan int)

	sp.mtx.Lock()
	defer sp.mtx.Unlock()

	sp.changes = append(sp.changes, c)

	return c
}

func (sp *s3Poll) delay() time.Duration {
	// TODO: configuration options?
	r := time.Duration(rand.Int31n(10)) * time.Second
	return (time.Second * 10) + r
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

func (sp *s3Poll) poll() error {
	client, err := sp.client()
	if err != nil {
		return err
	}

	bucket := client.Bucket(sp.bucket)

	log.WithFields(log.Fields{
		"last_etag": sp.lastEtag,
		"bucket":    sp.bucket,
		"file":      ".distsync",
	}).Debug("Checking for changed ETag")

	resp, err := bucket.Head(".distsync")

	if err != nil {
		return err
	}

	etag := resp.Header.Get("ETag")
	if etag == "" {
		return errors.New("Empty ETag on HEAD")
	}

	if etag != sp.lastEtag {
		log.WithFields(log.Fields{
			"last_etag": sp.lastEtag,
			"new_etag":  etag,
		}).Info("ETag changed, notifying watchers.")

		sp.lastEtag = etag
		sp.broadcast()
	}

	return nil
}

func (sp *s3Poll) mainLoop() {
	defer sp.wg.Done()

	// we use time.After instead of a ticker because
	// the requests to a bandend might take awhile (eg, re-authentication)
	// and we don't want to piss off cloud operators or our bill too much.
	timeChan := time.After(sp.delay())

	errCount := 0

	for {
		select {
		case <-timeChan:
			err := sp.poll()
			if err != nil {
				log.WithFields(log.Fields{
					"error":       err,
					"error_count": errCount,
				}).Error("Error while polling s3.")
				// TODO: configuration?
				if errCount > 10 {
					log.Error(">10 consecutive errors while polling, exiting.")
					// TODO: figure out better interface to Daemon code.
					os.Exit(1)
				}
			} else {
				errCount = 0
			}

			timeChan = time.After(sp.delay())
		case <-sp.quit:
			return
		}
	}
}

func (sp *s3Poll) Start() error {
	sp.wg.Add(1)
	go sp.mainLoop()
	return nil
}

func (sp *s3Poll) Stop() error {
	close(sp.quit)
	sp.wg.Wait()
	return nil
}
