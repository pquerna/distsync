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

package crypto

import (
	"github.com/codahale/chacha20poly1305"
	"github.com/codahale/etm"

	"bytes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
)

type EtmCryptor struct {
	ctype  string
	secret []byte
	c      cipher.AEAD
}

// 32-byte secret
func NewAES128SHA256(secret []byte) (Cryptor, error) {
	e, err := etm.NewAES128SHA256(secret)
	if err != nil {
		return nil, err
	}

	return &EtmCryptor{
		secret: secret,
		c:      e,
		ctype:  "AEAD_AES_128_CBC_HMAC_SHA_256",
	}, nil
}

// 32-byte secret
func NewChacha20poly1305(secret []byte) (Cryptor, error) {
	e, err := chacha20poly1305.New(secret)
	if err != nil {
		return nil, err
	}

	return &EtmCryptor{
		secret: secret,
		c:      e,
		ctype:  "AEAD_CHACHA20_POLY1305",
	}, nil
}

var v1header = []byte("distsync01")
var v1chunkSize = uint32(1000000)
var v1maxChunkSize = uint32(v1chunkSize * 10)

//
// Encrypts an cleartext input Reader in 1 megabyte chunks.
//
// File Format:
//
// Header: 10 bytes for version and cipher identification.
//		"distsync01": v1, AEAD_AES_128_CBC_HMAC_SHA_256.
//		"distsync02": v2, AEAD_CHACHA20_POLY1305
// Data block(s):
// 		4-bytes chunk size. (PutUint32)
// 		AEAD encrypted data. (up to `v1maxChunkSize`)
// Trailing hash block:
// 		0 byte data block, followed by:
//		mac []byte: 32 byte HMAC of file's contents.
func (e *EtmCryptor) Encrypt(r io.Reader, w io.Writer) error {
	buf := make([]byte, v1chunkSize)
	nonce := make([]byte, e.c.NonceSize())
	enbuf := make([]byte, cap(buf)+e.c.Overhead())
	lbuf := make([]byte, 4)
	// TOOD: TeeWriter for HMAC?
	mac := hmac.New(sha256.New, e.secret)

	s := ""
	switch e.ctype {
	case "AEAD_AES_128_CBC_HMAC_SHA_256":
		s = "distsync01"
	case "AEAD_CHACHA20_POLY1305":
		s = "distsync02"
	}

	_, err := io.WriteString(w, s)
	mac.Write([]byte(s))

	if err != nil {
		return err
	}

	for {
		enbuf = enbuf[0:0]

		n, err := r.Read(buf)

		if n > 0 {
			_, err := rand.Read(nonce)
			if err != nil {
				return err
			}

			enbuf = e.c.Seal(enbuf, nonce, buf[:n], []byte{})

			binary.BigEndian.PutUint32(lbuf, uint32(len(enbuf)))

			_, err = w.Write(lbuf)
			mac.Write(lbuf)
			if err != nil {
				return err
			}

			_, err = w.Write(enbuf)
			mac.Write(enbuf)
			if err != nil {
				return err
			}
		}

		if err == io.EOF {
			binary.BigEndian.PutUint32(lbuf, 0)
			_, err = w.Write(lbuf)
			mac.Write(lbuf)
			if err != nil {
				return err
			}
			break
		} else if err != nil {
			return err
		}
	}

	outmac := mac.Sum(nil)

	_, err = w.Write(outmac)
	if err != nil {
		return err
	}

	return nil
}

func (e *EtmCryptor) Decrypt(r io.Reader, w io.Writer) error {
	header := make([]byte, 10)
	lbuf := make([]byte, 4)
	mac := hmac.New(sha256.New, e.secret)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return err
	}

	if bytes.Compare(v1header, header) != 0 {
		return errors.New("Unknown header in encrypted file.")
	}

	mac.Write(header)

	for {
		_, err := io.ReadFull(r, lbuf)
		if err != nil {
			return err
		}

		mac.Write(lbuf)
		llen := binary.BigEndian.Uint32(lbuf)
		if llen > v1maxChunkSize {
			return errors.New("invalid size in of encrypted chunk")
		}

		if llen == 0 {
			// EOF, zero length block, next 32 bytes are HMAC.
			messageMAC := make([]byte, 32)
			_, err = io.ReadFull(r, messageMAC)
			if err != nil {
				return err
			}

			expectedMAC := mac.Sum(nil)
			rv := hmac.Equal(messageMAC, expectedMAC)
			if rv != true {
				return errors.New("File HMAC failed.")
			}
			return nil
		}

		buf := make([]byte, llen)
		clearbuf := make([]byte, 0, llen)

		_, err = io.ReadFull(r, buf)

		if err != nil {
			return err
		}

		mac.Write(buf)

		clearbuf, err = e.c.Open(clearbuf, nil, buf, []byte{})

		if err != nil {
			return err
		}

		_, err = w.Write(clearbuf)
		if err != nil {
			return err
		}
	}
}
