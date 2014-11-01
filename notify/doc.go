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

package notify

import (
	"github.com/pquerna/distsync/common"

	"errors"
	"strings"
)

// A notifier provides a channel for when
// the Backend storage area has changed.
// Callers should then call their Storage engine with List().
type Notifier interface {
	Start() error
	Stop() error
	// Channel gets a single item when something changes.
	Changed() chan int
}

func NewFromConf(c *common.Conf) (Notifier, error) {
	switch strings.ToUpper(c.Notify) {
	case "S3POLL":
		return NewS3Poll(c.Aws, c.StorageBucket)
	case "CLOUDFILESPOLL":
		return NewCloudFilesPoll(c.Rackspace, c.StorageBucket)
	}

	return nil, errors.New("Unknown Notify backend: " + c.Notify)
}
