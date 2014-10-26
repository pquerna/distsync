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

package common

import (
	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"

	"bytes"
	"io/ioutil"
)

type Conf struct {
	SharedSecret  string
	Encrypt       string
	Notify        string
	Storage       string
	StorageBucket string
	OutputDir     *string
	AwsCreds      *AwsCreds
	PeerDist      *PeerDist
}

type PeerDist struct {
	Region     string
	ListenAddr string
	GossipAddr string
}

type AwsCreds struct {
	Region    string
	AccessKey string
	SecretKey string
}

func NewConf() *Conf {
	return &Conf{
		Encrypt:   "AEAD_AES_128_CBC_HMAC_SHA_256",
		Notify:    "S3Poll",
		Storage:   "S3",
		OutputDir: nil,
		AwsCreds:  nil,
		PeerDist:  nil,
	}
}

func ConfFromFile(file string) (*Conf, error) {
	file, err := homedir.Expand(file)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	c := NewConf()

	_, err = toml.Decode(string(data), c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Conf) ToString() (string, error) {
	buf := bytes.Buffer{}
	err := toml.NewEncoder(&buf).Encode(c)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
