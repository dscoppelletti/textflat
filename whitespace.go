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

import "unicode"

type WhitespaceFilter struct {

	output chan rune
	lineLen int
	lastCharBlank bool
	lastCharCR bool
	lastLineEmpty bool
	pendingBlank bool
	pendingNL bool
}

func doWhitespaceFilter(input chan rune, output chan rune) {
	filter := newWhitespaceFilter(output)
	for r := range input {
		filter.onRune(r)
	}

	close(output)
}

func newWhitespaceFilter(output chan rune) *WhitespaceFilter {
	return &WhitespaceFilter { output: output, lineLen: 0, lastCharBlank: true,
		lastCharCR: false, lastLineEmpty: true, pendingBlank: false,
		pendingNL: false }
}

func (filter *WhitespaceFilter) onRune(r rune) {
	if r == NL {
		if filter.lastCharCR {
			// No NL after CR
			filter.lastCharCR = false
		} else {
			filter.onNL()
		}

		return
	}

	if r == CR {
		filter.onNL()

		// Maybe there will be NL after CR
		filter.onAfterCR()
		return
	}

	if unicode.IsSpace(r) {
		// No whitespace at the start of line
		if !filter.lastCharBlank {
			filter.pendingBlank = true
		}

		return
	}

	if filter.pendingNL {
		filter.addNL()
	}

	if filter.pendingBlank {
		filter.output <- BLANK
		filter.onAfterBlank()
	}

	filter.output <- r
	filter.onAfterRune()
}

func (filter *WhitespaceFilter) addNL() {
	filter.output <- NL
	filter.onAfterNL()
}

func (filter *WhitespaceFilter) onNL() {
	if filter.lineLen == 0 {
		// Will not insert NL after NL
		if !filter.lastLineEmpty {
			filter.pendingNL = true
		}
	} else {
		filter.addNL()
	}
}

func (filter *WhitespaceFilter) onAfterBlank() {
	filter.lineLen++
	filter.lastCharBlank = true
	filter.pendingBlank = false
}

func (filter *WhitespaceFilter) onAfterCR() {
	filter.lastCharCR = true
}

func (filter *WhitespaceFilter) onAfterNL() {
	filter.lastCharBlank = true
	filter.lastCharCR = false
	filter.lastLineEmpty = (filter.lineLen == 0)
	filter.lineLen = 0
	filter.pendingBlank = false
	filter.pendingNL = false
}

func (filter *WhitespaceFilter) onAfterRune() {
	filter.lineLen++
	filter.lastCharBlank = false
	filter.lastCharCR = false
}
