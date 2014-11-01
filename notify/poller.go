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

	"math/rand"
	"os"
	"sync"
	"time"
)

type timedPoller struct {
	mtx     sync.Mutex
	wg      sync.WaitGroup
	changes []chan int
	quit    chan int
	poller  Poller
}

type Poller interface {
	Poll() (bool, error)
}

func newTimedPoller(p Poller) *timedPoller {
	return &timedPoller{
		changes: make([]chan int, 0),
		quit:    make(chan int),
		poller:  p,
	}
}

func (p *timedPoller) broadcast() {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, v := range p.changes {
		v <- 1
	}
}

func (p *timedPoller) Changed() chan int {
	c := make(chan int)

	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.changes = append(p.changes, c)

	return c
}

func (p *timedPoller) delay() time.Duration {
	// TODO: configuration options?
	r := time.Duration(rand.Int31n(10)) * time.Second
	return (time.Second * 10) + r
}

func (p *timedPoller) mainLoop() {
	defer p.wg.Done()

	// we use time.After instead of a ticker because
	// the requests to a bandend might take awhile (eg, re-authentication)
	// and we don't want to piss off cloud operators or our bill too much.
	timeChan := time.After(p.delay())

	errCount := 0

	for {
		select {
		case <-timeChan:
			changed, err := p.poller.Poll()
			if err != nil {
				log.WithFields(log.Fields{
					"error":       err,
					"error_count": errCount,
				}).Error("Error while polling.")
				// TODO: configuration?
				if errCount > 10 {
					log.Error(">10 consecutive errors while polling, exiting.")
					// TODO: figure out better interface to Daemon code.
					os.Exit(1)
				}
			} else {
				errCount = 0
				if changed {
					p.broadcast()
				}
			}

			timeChan = time.After(p.delay())
		case <-p.quit:
			return
		}
	}
}

func (p *timedPoller) Start() error {
	p.wg.Add(1)
	go p.mainLoop()
	return nil
}

func (p *timedPoller) Stop() error {
	close(p.quit)
	p.wg.Wait()
	return nil
}
