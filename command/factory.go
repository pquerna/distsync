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

package command

import (
	"github.com/mitchellh/cli"
)

func Factory(ui cli.Ui) map[string]cli.CommandFactory {
	x := map[string]cli.CommandFactory{
		"download": func() (cli.Command, error) {
			return &Download{
				Ui: ui,
			}, nil
		},
		"daemon": func() (cli.Command, error) {
			return &Daemon{
				Ui: ui,
			}, nil
		},
		"upload": func() (cli.Command, error) {
			return &Upload{
				Ui: ui,
			}, nil
		},
		"setup": func() (cli.Command, error) {
			return &Setup{
				Ui: ui,
			}, nil
		},
	}
	return x
}
