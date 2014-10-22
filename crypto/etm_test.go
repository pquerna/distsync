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
	"bytes"
	"strings"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	src := bytes.NewReader([]byte("hello world"))
	dst := &bytes.Buffer{}
	roundtrip := &bytes.Buffer{}

	ec, err := NewEtmCryptor([]byte("hellohelloworld1hellohelloworld1"))
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	err = ec.Encrypt(src, dst)

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	enreader := bytes.NewReader(dst.Bytes())
	err = ec.Decrypt(enreader, roundtrip)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	st := roundtrip.String()
	if st != "hello world" {
		t.Fatal("Failed round trip.")
	}
}

func TestTampered(t *testing.T) {
	src := bytes.NewReader([]byte("hello world"))
	dst := &bytes.Buffer{}
	roundtrip := &bytes.Buffer{}

	ec, err := NewEtmCryptor([]byte("hellohelloworld1hellohelloworld1"))
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	err = ec.Encrypt(src, dst)

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	b := dst.Bytes()
	b[43] += 1

	enreader := bytes.NewReader(b)
	err = ec.Decrypt(enreader, roundtrip)
	if err != nil {
		return
	}
	t.Fatalf("Missing error from tampered data: enreader:%v", enreader)

}

func TestEtmName(t *testing.T) {
	ec, err := NewEtmCryptor([]byte("hellohelloworld1hellohelloworld1"))
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	fname := "hello-world-1.2.3.4.tar.gz"

	s, err := ec.EncryptName(fname)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	fname2, err := ec.DecryptName(s)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if fname != fname2 {
		t.Fatalf("error: expected decrypted name of %s, but got %s", fname, fname2)
	}

	tampered := strings.ToUpper(s)

	fname2, err = ec.DecryptName(tampered)
	if err == nil {
		t.Fatalf("expected error due to tampering, didn't get one.")
	}
}
