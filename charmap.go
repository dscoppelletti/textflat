// Copyright (C) 2022 Dario Scoppelletti, <http://www.scoppelletti.it/>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
)

func doCharMapFilter(input chan rune, output chan rune,
	charMap map[string]string) {
	for r := range input {
		key := fmt.Sprintf("%04X", r)
		if value, ok := charMap[key]; ok {
			for _, c := range value {
				output <- c
			}
		} else {
			output <- r
		}
	}

	close(output)
}

type flat struct {
	Value string
	Xmlencoded bool
}

func loadCharMap(name string) (map[string]string, error) {
	in, err := os.Open(name)
		if err != nil {
			return nil, err
		}

	defer in.Close()
	dec := json.NewDecoder(in)

	var m1 map[string]flat
	if err := dec.Decode(&m1); err != nil {
		return nil, err
	}

	m2 := make(map[string]string, len(m1))
	for key, flat := range m1 {
		value := flat.Value
		if flat.Xmlencoded {
			value = html.UnescapeString(value)
		}

		m2[key] = value
	}

	return m2, nil
}
