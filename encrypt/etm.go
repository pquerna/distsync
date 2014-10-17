package encrypt

import (
	"github.com/codahale/etm"

	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
)

type EtmCryptor struct {
	c     cipher.AEAD
	nonce []byte
}

// TODO: make streaming :*(
// 32-byte key
func NewEtmCryptor(secret []byte) (Cryptor, error) {
	e, err := etm.NewAES128SHA256(secret)
	if err != nil {
		return nil, err
	}

	return &EtmCryptor{
		c:     e,
		nonce: nil,
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
// Data block:
// 		4-bytes chunk size. (PutUint32)
// 		AEAD encrypted data.
func (e *EtmCryptor) Encrypt(r io.Reader, w io.Writer) error {
	buf := make([]byte, v1chunkSize)
	nonce := make([]byte, e.c.NonceSize())
	enbuf := make([]byte, int(v1chunkSize)+e.c.Overhead())
	io.WriteString(w, "distsync01")

	for {
		buf = buf[0:0]
		enbuf = enbuf[0:0]
		nonce = nonce[0:0]

		n, err := r.Read(buf)

		if n > 0 {
			_, err := rand.Read(nonce)
			if err != nil {
				return err
			}

			_ = e.c.Seal(enbuf, nonce, buf[:n], []byte{})

			lbuf := make([]byte, 4)
			binary.BigEndian.PutUint32(lbuf, uint32(len(enbuf)))

			_, err = w.Write(lbuf)
			if err != nil {
				return err
			}

			_, err = w.Write(enbuf)
			if err != nil {
				return err
			}
		}

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (e *EtmCryptor) Decrypt(r io.Reader, w io.Writer) error {
	header := make([]byte, 10)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return err
	}

	if bytes.Compare(v1header, header) != 0 {
		return errors.New("Unknown header in encrypted file.")
	}

	for {
		lbuf := make([]byte, 4)
		_, err := io.ReadFull(r, lbuf)
		if err == io.ErrUnexpectedEOF {
			return nil
		} else if err != nil {
			return err
		}

		llen := binary.BigEndian.Uint32(lbuf)
		if llen > v1maxChunkSize {
			return errors.New("invalid size in of encrypted chunk")
		}

		buf := make([]byte, llen)
		clearbuf := make([]byte, llen)

		_, err = io.ReadFull(r, buf)

		if err != nil {
			return err
		}

		_, err = e.c.Open(clearbuf, nil, buf, []byte{})
		if err != nil {
			return err
		}

		_, err = w.Write(clearbuf)
		if err != nil {
			return err
		}
	}
}
