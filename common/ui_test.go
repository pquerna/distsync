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
	"testing"
	"time"
)

func TestHumanizeRate(t *testing.T) {
	// TOOD: more test cases
	s := HumanizeRate(100000*99, time.Second*5)
	if s != "2.0MB/s" {
		t.Fatal("HumanizeRate: expected 2.0MB/s, got " + s)
	}
}
