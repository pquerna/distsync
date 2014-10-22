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
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"hash/crc32"
)

func RandomSecret() (string, error) {
	crcbuf := make([]byte, 4)
	buf := make([]byte, 32)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	crc := crc32.ChecksumIEEE(buf)

	binary.BigEndian.PutUint32(crcbuf, crc)

	return hex.EncodeToString(append(buf, crcbuf...)), nil
}

func decodeSecret(secin string, seclen int) ([]byte, error) {
	if len(secin) != hex.EncodedLen(seclen+4) {
		return nil, errors.New("Invalid shared secret, length is wrong?")
	}

	buf, err := hex.DecodeString(secin)
	if err != nil {
		return nil, err
	}

	crca := crc32.ChecksumIEEE(buf[:seclen])

	crcb := binary.BigEndian.Uint32(buf[seclen:])

	if crca != crcb {
		return nil, errors.New("SharedSecret failed CRC check.")
	}

	return buf[:seclen], nil
}
