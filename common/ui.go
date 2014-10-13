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
	"github.com/mitchellh/cli"

	"fmt"
	"strconv"
	"strings"
)

func Choice(ui cli.Ui, question string, options []string) (int, error) {

	ui.Output("Select option:")

	for i := 0; i < len(options); i++ {
		ui.Output(fmt.Sprintf("  %d) %s", i+1, options[i]))
	}

	if len(options) == 1 {
		ui.Output(question)
		ui.Output(fmt.Sprintf("Selected %s", options[0]))
		return 0, nil
	}

	for {
		answer, err := ui.Ask(question + fmt.Sprintf(" [1-%d]", len(options)))
		if err != nil {
			return 0, err
		}

		if v, err := strconv.Atoi(answer); err == nil {
			s := v - 1
			if s > len(options) || s < 0 {
				ui.Output(fmt.Sprintf("Invalid selection: %d", v))
				continue
			}
			return s, nil
		}

		for i, s := range options {
			if strings.ToLower(s) == strings.ToLower(answer) {
				return i, nil
			}
		}

		ui.Output(fmt.Sprintf("Invalid selection: %s", answer))
		continue
	}
}
