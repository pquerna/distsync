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
	"github.com/pquerna/distsync/common"

	"sync"
)

type s3Poll struct {
	mtx     sync.Mutex
	wg      sync.WaitGroup
	bucket  string
	conf    *common.AwsCreds
	changes []chan int
	quit    chan int
}

// Polls the specified S3 bucket for a new manifest file every 10 seconds.
// $0.004 per 10,000 requests = 0.10368 per month per watcher.
func NewS3Poll(conf *common.AwsCreds, bucketName string) (Notifier, error) {
	return &s3Poll{
		bucket:  bucketName,
		conf:    conf,
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

func (sp *s3Poll) mainLoop() {
	defer sp.wg.Done()
	for {
		select {
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
