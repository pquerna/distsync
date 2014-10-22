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

package encrypt

import (
	"github.com/pquerna/distsync/common"

	"errors"
	"io"
)

type Encryptor interface {
	EncryptName(name string) (string, error)
	Encrypt(io.Reader, io.Writer) error
}

type Decryptor interface {
	DecryptName(name string) (string, error)
	Decrypt(io.Reader, io.Writer) error
}

type Cryptor interface {
	Encryptor
	Decryptor
}

func NewFromConf(c *common.Conf) (Cryptor, error) {
	// currently etm use a 32 byte secret.
	// TODO: better abstraction / interface.
	secret, err := decodeSecret(c.SharedSecret, 32)
	if err != nil {
		return nil, err
	}

	switch c.Encrypt {
	case "AEAD_AES_128_CBC_HMAC_SHA_256":
		return NewEtmCryptor(secret)
	}

	return nil, errors.New("Unknown crypto backend: " + c.Encrypt)
}
