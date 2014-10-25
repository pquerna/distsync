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
	_ "github.com/VividCortex/ewma" // TDOO: find better library.

	"sync"
)

type PeerStatus int

const (
	PEER_STATUS_UNKNOWN PeerStatus = 1 << iota
	PEER_STATUS_OK
	PEER_STATUS_DEAD
)

type PeerRange int

const (
	PEER_RANGE_UNKNOWN PeerRange = 1 << iota
	PEER_RANGE_LOCAL
	PEER_RANGE_REMOTE
)

type Peer interface {
	Id() string
	Status() PeerStatus
	Range() PeerRange

	// Latency in seconds
	Latency() float64

	Start() error
	Stop() error
}

func (p PeerStatus) String() string {
	switch p {
	case PEER_STATUS_UNKNOWN:
		return "UNKNOWN"
	case PEER_STATUS_OK:
		return "OK"
	case PEER_STATUS_DEAD:
		return "DEAD"
	}
	panic("unreached")
}

func (p PeerRange) String() string {
	switch p {
	case PEER_RANGE_UNKNOWN:
		return "UNKNOWN"
	case PEER_RANGE_LOCAL:
		return "LOCAL"
	case PEER_RANGE_REMOTE:
		return "REMOTE"
	}
	panic("unreached")
}

type RemotePeer struct {
	mtx       sync.Mutex
	wg        sync.WaitGroup
	quit      chan int
	id        string
	status    PeerStatus
	addresses []string
}

func NewPeer(id string, addresses []string) (Peer, error) {
	return &RemotePeer{
		quit: make(chan int),
	}, nil
}

func (rp *RemotePeer) Id() string {
	return rp.id
}

func (rp *RemotePeer) Status() PeerStatus {
	rp.mtx.Lock()
	defer rp.mtx.Unlock()

	if rp.Id() == "" {
		return PEER_STATUS_UNKNOWN
	}

	return rp.status
}

func (rp *RemotePeer) Latency() float64 {
	rp.mtx.Lock()
	defer rp.mtx.Unlock()

	//	return latency.Value()
	return 0.0
}

func (rp *RemotePeer) Start() error {
	rp.wg.Add(1)
	return nil
}

func (rp *RemotePeer) Stop() error {
	close(rp.quit)
	rp.wg.Wait()
	return nil
}

func (rp *RemotePeer) Range() PeerRange {
	if rp.Status() != PEER_STATUS_OK {
		return PEER_RANGE_UNKNOWN
	}

	dist := rp.Latency()

	// TODO: actual algorithms here.
	if dist > 0.025 {
		return PEER_RANGE_REMOTE
	} else {
		return PEER_RANGE_LOCAL
	}
}
