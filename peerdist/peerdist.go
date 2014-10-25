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
	"github.com/pquerna/distsync/crypto"

	"io"
	"sync"
)

type FileInfo struct {
	name      string
	humanName string
	length    int64
	origin    RangeDownloader
}

type RangeDownloader interface {
	RangeDownload(name string, offset int64, length int64, writer io.Writer) error
}

type Service interface {
	Start() error
	Stop() error
}

type DistDownloader interface {
	Service
	Download(fi *FileInfo, writer io.Writer) error
}

type peerDist struct {
	wg         sync.WaitGroup
	quit       chan int
	id         string
	secret     []byte
	listenAddr string
	gossipAddr string
}

func New(secret []byte, listenAddr string, gossipAddr string, d Discoverer) (DistDownloader, error) {

	id, err := crypto.RandomThing(32, false)

	if err != nil {
		return nil, err
	}

	return &peerDist{
		quit:       make(chan int),
		id:         id,
		secret:     secret,
		listenAddr: listenAddr,
		gossipAddr: gossipAddr,
	}, nil
}

func (pd *peerDist) Download(fi *FileInfo, writer io.Writer) error {
	return nil
}

func (pd *peerDist) Id() string {
	return pd.id
}

func (pd *peerDist) listen() error {
	return nil
}

func (pd *peerDist) Start() error {
	err := pd.listen()
	if err != nil {
		return err
	}
	pd.wg.Add(1)
	return nil
}

func (pd *peerDist) Stop() error {
	close(pd.quit)
	pd.wg.Wait()
	return nil
}
