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
)

type s3Poll struct {
	conf *common.S3Conf
}

// Polls the specified S3 bucket for a new manifest file every 10 seconds.
// $0.004 per 10,000 requests = 0.10368 per month per watcher.
func NewS3Poll(conf *common.S3Conf) Notifier {
	return nil
}
