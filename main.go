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
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
)

const (
	CR = rune('\r')
	NL = rune('\n')
	BLANK = rune(' ')
)

var inputFile string
var outputFile string
var overwrite bool
var collapseSpace bool
var lineLenMax int
var keepNL bool
var charMapFile string

func init() {
	flag.StringVar(&inputFile, "input", "", "Input file (default stdin).")
	flag.StringVar(&outputFile, "output", "", "Output file (default stdout).")
	flag.BoolVar(&overwrite, "overwrite", false,
		"Whether output file can be overwritten (default false).")
	flag.BoolVar(&collapseSpace, "collapsespace", false,
		"Whether the whitespace character sequences have to be collapsed in " +
		"a single space character (default false).")
	flag.IntVar(&lineLenMax, "wordwrap", 0,
		"Word-wrap lines at the specified column (default 0 = no word-wrap).")
	flag.BoolVar(&keepNL, "keepnewline", false,
		"Whether the original newline characters have to be kept (default " +
		"false).")
	flag.StringVar(&charMapFile, "charmap", "",
		"Character replacing map file (default no character replacing).")
}

func main() {
	flag.Parse()
	err := checkFlags()
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%v\n", err)
		flag.PrintDefaults()
		os.Exit(2)
	}

	var charMap = map[string]string { }
	if charMapFile != "" {
		m, err := loadCharMap(charMapFile)
		if err != nil {
			panic(err)
		}

		charMap = m
	}

	var reader *bufio.Reader
	var writer *bufio.Writer

	if inputFile == "" {
		reader = bufio.NewReader(os.Stdin)
	} else {
		in, err := os.Open(inputFile)
		if err != nil {
			panic(err)
		}

		defer in.Close()
		reader = bufio.NewReader(in)
	}

	if outputFile == "" {
		writer = bufio.NewWriter(os.Stdout)
	} else {
		var mode int
		if overwrite {
			mode = os.O_TRUNC
		} else {
			mode = os.O_EXCL
		}

		out, err := os.OpenFile(outputFile, os.O_WRONLY | os.O_CREATE | mode,
			0644)
		if err != nil {
			panic(err)
		}

		defer out.Close()
		writer = bufio.NewWriter(out)
	}
	defer writer.Flush()

	var input chan rune
	output := make(chan rune, 10)

	go readFile(reader, output)

	if charMapFile != "" {
		input = output
		output = make(chan rune)

		go doCharMapFilter(input, output, charMap)
	}

	if collapseSpace {
		input = output
		output = make(chan rune)

		go doWhitespaceFilter(input, output)
	}

	if lineLenMax > 0 {
		input = output
		output = make(chan rune)

		go doWordWrapFilter(input, output, lineLenMax, keepNL)
	}

	input = output
	writeFile(input, writer)
}

func checkFlags() error {
	if overwrite && outputFile == "" {
		return errors.New("Flag -overwrite is invalid without flag -output.")
	}
	if keepNL && lineLenMax <= 0 {
		return errors.New(
			"Flag -keepnewline is invalid without flag -wordwrap.")
	}

	return nil
}
