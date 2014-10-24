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
	"github.com/jackpal/Taipei-Torrent/torrent"
	"github.com/mitchellh/go-homedir"
	"github.com/pquerna/distsync/common"
	"github.com/pquerna/distsync/crypto"

	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type TorrentDownloader struct {
	storage      Storage
	ec           crypto.Cryptor
	torrentDir   string
	localDir     string
	wg           sync.WaitGroup
	quitChan     chan bool
	torrentFlags torrent.TorrentFlags
	torrents     map[string]*torrent.TorrentSession
	torrentConns chan *torrent.BtConn
}

func NewTorrentDownloader(conf *common.Conf) (*TorrentDownloader, error) {
	localDir, err := homedir.Expand(conf.LocalDirectory)
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
	tdir := path.Join(localDir, ".distsync-torrents")

	return &TorrentDownloader{
		storage:    st,
		ec:         ec,
		torrentDir: tdir,
		localDir:   localDir,
		torrentFlags: torrent.TorrentFlags{
			// TODO: dynamic listening port.
			Port:    6881,
			FileDir: tdir,
			// TODO: patch Taipei to support 'seed forever' mode.
			SeedRatio: 9000.0,
			// TODO: configuration options for NAT/UPNP/etc
		},
	}, nil
}

func (td *TorrentDownloader) Download(filename string, writer io.Writer) error {
	err := os.MkdirAll(td.torrentDir, 0755)
	if err != nil {
		return err
	}

	tmpTorrent, err := ioutil.TempFile(td.torrentDir, ".distsync-t")
	if err != nil {
		return err
	}
	defer func() {
		tmpTorrent.Close()
		os.Remove(tmpTorrent.Name())
	}()

	tdl, ok := td.storage.(DownloadTorrenter)
	if ok {
		err = tdl.DownloadTorrent(filename, tmpTorrent)
	} else {
		// TODO: this isn't a real thing. hrm.
		err = td.storage.Download(filename+".torrent", tmpTorrent)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to download .torrent file from origin.")
		return err
	}

	destname := path.Join(td.torrentDir, filename+".torrent")

	err = os.Rename(tmpTorrent.Name(), destname)
	if err != nil {
		return err
	}

	return errors.New("not implemented")
}

func (td *TorrentDownloader) Start() error {
	err := td.loadExistingTorrents()
	if err != nil {
		return err
	}

	conns, listenPort, err := torrent.ListenForPeerConnections(&td.torrentFlags)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"port":  td.torrentFlags.Port,
		}).Error("Failed to listen for torrent connections.")
		return err
	}

	td.torrentConns = conns

	log.WithFields(log.Fields{
		"port": listenPort,
	}).Info("Listening for BitTorrent peers")

	td.wg.Add(1)
	go td.mainLoop()
	return nil
}

func (td *TorrentDownloader) Stop() error {
	close(td.quitChan)
	td.wg.Wait()
	return nil
}

func (td *TorrentDownloader) torrentNameToClear(tf string) (string, error) {
	_, tf = filepath.Split(tf)
	tf = strings.TrimSuffix(tf, ".torrent")
	name, err := td.ec.DecryptName(tf)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (td *TorrentDownloader) loadExistingTorrents() error {
	matches, err := filepath.Glob(td.torrentDir + "*.torrent")
	if err != nil {
		return err
	}

	for _, tf := range matches {
		tname, err := td.torrentNameToClear(tf)
		if err != nil {
			log.WithFields(log.Fields{
				"error":        err,
				"torrent_file": tf,
			}).Error("Failed to decrypt name of existing torrent from work directory.")
			return err
		}

		ts, err := torrent.NewTorrentSession(
			&td.torrentFlags,
			tf,
			uint16(td.torrentFlags.Port))

		if err != nil {
			log.WithFields(log.Fields{
				"error":        err,
				"torrent_file": tf,
				"name":         tname,
			}).Error("Failed to load existing torrent from work directory.")
			return err
		}

		log.WithFields(log.Fields{
			"infohash":     ts.M.InfoHash,
			"torrent_file": tf,
			"name":         tname,
		}).Info("Starting torrent session")

		td.torrents[ts.M.InfoHash] = ts
	}

	return nil
}

func (td *TorrentDownloader) mainLoop() {
	defer td.wg.Done()
	for {
		select {

		case c := <-td.torrentConns:
			log.WithFields(log.Fields{
				"infohash":     c.Infohash,
				"peer_address": c.RemoteAddr,
			}).Info("New BitTorrent peer connection")

			ts, ok := td.torrents[c.Infohash]
			if ok {
				ts.AcceptNewPeer(c)
			} else {
				log.WithFields(log.Fields{
					"infohash":     c.Infohash,
					"peer_address": c.RemoteAddr.String(),
				}).Warn("Peer connected regarding unknown torrent.")
			}

		case <-td.quitChan:
			for k, ts := range td.torrents {
				err := ts.Quit()
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
						"name":  k,
					}).Error("Failed to stop torrent")
				} else {
					log.WithFields(log.Fields{
						"name": k,
					}).Info("Stopped torrent")
				}
			}

		}
	}
}
