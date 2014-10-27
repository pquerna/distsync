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

	"io/ioutil"
	"os"
	"path"
	"sync"
)

type DownloadQueue struct {
	mtx        sync.Mutex
	wg         sync.WaitGroup
	quit       chan int
	concurrent int
	work       chan *FileDownload
	dl         Downloader
}

type FileDownload struct {
	wg       sync.WaitGroup
	FileInfo *FileInfo
	conf     *common.Conf
	Error    error
	done     chan *FileDownload
}

// TODO: interface? meh.
func NewDownloadQueue(dl Downloader) *DownloadQueue {
	return &DownloadQueue{
		dl:         dl,
		quit:       make(chan int),
		work:       make(chan *FileDownload),
		concurrent: 3,
	}
}

func (fd *FileDownload) Done(err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"file":  fd.FileInfo.Name,
			"error": err,
		}).Error("Download failed")
	} else {
		log.WithFields(log.Fields{
			"file": fd.FileInfo.Name,
		}).Info("Download complete")
	}

	fd.Error = err
	fd.wg.Done()
	fd.done <- fd
}

func (fd *FileDownload) Start() {
	fd.wg.Add(1)
}

func (fd *FileDownload) Stop() error {
	fd.wg.Wait()
	return nil
}

func (dq *DownloadQueue) Add(conf *common.Conf, fi *FileInfo, dchan chan *FileDownload) *FileDownload {
	dq.mtx.Lock()
	defer dq.mtx.Unlock()

	fd := &FileDownload{
		FileInfo: fi,
		conf:     conf,
		done:     dchan,
	}

	fd.wg.Add(1)

	dq.work <- fd

	return fd
}

func (dq *DownloadQueue) download(fd *FileDownload) error {
	var err error
	defer fd.Done(err)

	ec, err := crypto.NewFromConf(fd.conf)

	if err != nil {
		log.WithFields(log.Fields{
			"file":  fd.FileInfo.Name,
			"error": err,
		}).Error("Crypto setup failed")
		return err
	}

	workDir, err := homedir.Expand(*fd.conf.OutputDir)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  *fd.conf.OutputDir,
		}).Error("Home directory expansion failed")
		return err
	}

	finalName := path.Join(workDir, fd.FileInfo.Name)

	tmpFileEnc, err := ioutil.TempFile(workDir, ".distsync-e")
	if err != nil {
		log.WithFields(log.Fields{
			"file":    fd.FileInfo.Name,
			"workdir": workDir,
			"error":   err,
		}).Error("Failed to create temp file")
		return err
	}

	defer func() {
		tmpFileEnc.Close()
		os.Remove(tmpFileEnc.Name())
	}()

	err = dq.dl.Download(fd.FileInfo.EncryptedName, tmpFileEnc)
	if err != nil {
		log.WithFields(log.Fields{
			"file":           fd.FileInfo.Name,
			"encrypted_name": fd.FileInfo.EncryptedName,
			"workdir":        workDir,
			"error":          err,
		}).Error("Download failed")
		return err
	}

	tmpFile, err := ioutil.TempFile(workDir, ".distsync")
	if err != nil {
		log.WithFields(log.Fields{
			"file":    fd.FileInfo.Name,
			"workdir": workDir,
			"error":   err,
		}).Error("Failed to create temp file")
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	_, err = tmpFileEnc.Seek(0, 0)
	if err != nil {
		log.WithFields(log.Fields{
			"file":    fd.FileInfo.Name,
			"workdir": workDir,
			"error":   err,
		}).Error("Failed seek to 0 on temp file")
		return err
	}

	err = ec.Decrypt(tmpFileEnc, tmpFile)
	if err != nil {
		log.WithFields(log.Fields{
			"file":    fd.FileInfo.Name,
			"workdir": workDir,
			"error":   err,
		}).Error("Decryption failed.")
		return err
	}

	err = os.Chtimes(tmpFile.Name(), fd.FileInfo.LastModified, fd.FileInfo.LastModified)
	if err != nil {
		log.WithFields(log.Fields{
			"file":          fd.FileInfo.Name,
			"workdir":       workDir,
			"error":         err,
			"last_modified": fd.FileInfo.LastModified,
		}).Error("Failed to update modified time on file.")
		return err
	}

	err = os.Rename(tmpFile.Name(), finalName)
	if err != nil {
		log.WithFields(log.Fields{
			"soruce":  tmpFile.Name(),
			"dest":    fd.FileInfo.Name,
			"workdir": workDir,
			"error":   err,
		}).Error("Failed to rename file")
		return err
	}

	err = os.Chtimes(finalName, fd.FileInfo.LastModified, fd.FileInfo.LastModified)
	if err != nil {
		log.WithFields(log.Fields{
			"file":          finalName,
			"workdir":       workDir,
			"error":         err,
			"last_modified": fd.FileInfo.LastModified,
		}).Error("Failed to update modified time on file.")
		return err
	}

	return nil
}

func (dq *DownloadQueue) worker() {
	defer dq.wg.Done()

	for {
		select {
		case <-dq.quit:
			return
		case req := <-dq.work:
			err := dq.download(req)
			if err != nil {
				log.WithFields(log.Fields{
					"error":          err,
					"name":           req.FileInfo.Name,
					"encrypted_name": req.FileInfo.EncryptedName,
				}).Error("Error downloading.")
			}
		}
	}
}

func (dq *DownloadQueue) Start() error {
	for i := 0; i < dq.concurrent; i++ {
		dq.wg.Add(1)
		go dq.worker()
	}
	return nil
}

func (dq *DownloadQueue) Stop() error {
	close(dq.quit)
	dq.wg.Wait()
	return nil
}
