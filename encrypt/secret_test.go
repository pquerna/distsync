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

	"testing"
)

func TestSecrets(t *testing.T) {
	sec, err := RandomSecret()
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	c := common.Conf{
		SharedSecret: sec,
	}
	_, err = NewFromConf(&c)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	c = common.Conf{
		SharedSecret: "aa" + sec,
	}
	_, err = NewFromConf(&c)
	if err == nil {
		t.Fatal("expected error from corrupted secret")
	}

}
