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

type WordWrapFilter struct {

	output chan rune
	lineLenMax int
	keepNL bool
	lastCharCR bool
	line []rune
}

func doWordWrapFilter(input chan rune, output chan rune, lineLenMax int,
	keepNL bool) {
	filter := NewWordWrapFilter(output, lineLenMax, keepNL)
	for r := range input {
		filter.onRune(r)
	}

	close(output)
}

func NewWordWrapFilter(output chan rune, lineLenMax int,
	keepNL bool) *WordWrapFilter {
	return &WordWrapFilter { output: output, lineLenMax: lineLenMax,
		keepNL: keepNL, lastCharCR: false, line: make([]rune, 0, lineLenMax) }
}

func (filter *WordWrapFilter) onRune(r rune) {
	if r == NL {
		if filter.lastCharCR {
			// No NL after CR
			filter.lastCharCR = false
		} else if filter.keepNL {
			filter.flushLine()
		} else {
			filter.addBlank()
		}

		return
	}

	if r == CR {
		if filter.keepNL {
			filter.flushLine()
		} else {
			filter.addBlank()
		}

		// Maybe there will be NL after CR
		filter.lastCharCR = true
	}

	filter.lastCharCR = false
	if unicode.IsSpace(r) {
		filter.addBlank()
	} else {
		filter.addRune(r)
	}
}

func (filter *WordWrapFilter) addBlank() {
	n := len(filter.line)
	if n < filter.lineLenMax {
		// Enough space in current line
		if n > 0 {
			filter.line = append(filter.line, BLANK)
		}
	} else {
		// Current line exhausted:
		// Replace whitespace by NL.
		filter.flushLine()
	}
}

func (filter *WordWrapFilter) addRune(r rune) {
	if len(filter.line) < filter.lineLenMax {
		// Enough space in current line
		filter.line = append(filter.line, r)
		return
	}

	// Current line exhausted:
	// Search for the last whitespace where I can break the line.
	k := - 1
	for i := len(filter.line) - 1; i >= 0; i-- {
		if unicode.IsSpace(filter.line[i]) {
			k = i
			break
		}
	}

	if k < 0 {
		// No whitespace found:
		// Break the line anyway.
		filter.flushLine()
		filter.line = append(filter.line, r)
		return
	}

	// Flush the line until the last whitespace
	for _, c := range filter.line[0:k] {
		filter.output <- c
	}

	filter.output <- NL
	filter.line = append(filter.line[k + 1:], r)
}

func (filter *WordWrapFilter) flushLine() {
	for _, c := range filter.line {
		filter.output <- c
	}

	filter.output <- NL
	filter.line = make([]rune, 0, filter.lineLenMax)
}
