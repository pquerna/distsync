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
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/go-homedir"
	"github.com/pquerna/distsync/common"
	"github.com/pquerna/distsync/crypto"
	"github.com/pquerna/distsync/peerdist"

	"errors"
	"io"
)

type PeerDownloader struct {
	storage   Storage
	pd        peerdist.DistDownloader
	ec        crypto.Cryptor
	outputDir string
}

func NewPeerDownloader(conf *common.Conf) (PersistentDownloader, error) {
	if conf.OutputDir == nil {
		return nil, errors.New("Config Error: OutputDir must be set.")
	}

	outputDir, err := homedir.Expand(*conf.OutputDir)
	if err != nil {
		return nil, err
	}

	ec, err := crypto.NewFromConf(conf)
	if err != nil {
		return nil, err
	}

	// TODO: better separation? different interface for download?
	st, err := NewFromConf(conf)
	if err != nil {
		return nil, err
	}

	s3d, err := peerdist.NewS3Discovery(conf)
	if err != nil {
		return nil, err
	}

	la := ":4166"
	ga := ":4167"

	if conf.PeerDist != nil {
		if conf.PeerDist.ListenAddr != "" {
			la = conf.PeerDist.ListenAddr
		}
		if conf.PeerDist.GossipAddr != "" {
			ga = conf.PeerDist.GossipAddr
		}
	}

	pd, err := peerdist.New(nil, la, ga, s3d)
	if err != nil {
		return nil, err
	}

	return &PeerDownloader{
		pd:        pd,
		storage:   st,
		ec:        ec,
		outputDir: outputDir,
	}, nil
}

func (pd *PeerDownloader) Start() error {
	return nil
}

func (pd *PeerDownloader) Stop() error {
	return nil
}

func (pd *PeerDownloader) Download(filename string, writer io.Writer) error {
	log.Error("not downloading yet.")
	return nil
}
