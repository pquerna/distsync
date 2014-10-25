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

package peerdist

import (
	"github.com/pquerna/distsync/common"

	"sync"
)

type Discoverer interface {
	Service
	Advertise(gossipAddr string)
	Peers() []string
}

type s3Discovery struct {
	wg   sync.WaitGroup
	quit chan int
}

func NewS3Discovery(conf *common.Conf) (Discoverer, error) {
	return &s3Discovery{
		quit: make(chan int),
	}, nil
}

func (sd *s3Discovery) doAdvertise(gossipAddr string) {

}

func (sd *s3Discovery) Advertise(gossipAddr string) {
	sd.wg.Add(1)
	sd.doAdvertise(gossipAddr)
}

func (sd *s3Discovery) Peers() []string {
	return nil
}

func (sd *s3Discovery) Start() error {
	return nil
}

func (sd *s3Discovery) Stop() error {
	close(sd.quit)
	sd.wg.Wait()
	return nil
}
