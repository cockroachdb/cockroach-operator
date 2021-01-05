/*
Copyright 2021 The Cockroach Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testutil

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// ValidateHeaders represents a project that that needs to be scanned.
type ValidateHeaders struct {
	// A slice of fileNames to check.  If this value does not exist all
	// files under the rootDir are checked.
	fileNames *[]string
	// Root directory of the project to check.
	rootDir string

	// Directory that containes the boilerplate and preamble files.
	boilerplateDir string
	// Force the checking of a specfic file extension.
	forceExtension string

	// Used to store the boilerplate files.
	references *map[string][]string
	// Used to store the preamble files.
	preambles *map[string][]string
}

var DateRegex = "20\\d\\d"

var SkippedPaths = map[string]bool{
	"Godeps":      true,
	"third_party": true,
	"_gopath":     true,
	"_output":     true,
	".git":        true,
	"vendor":      true,
	"external":    true,
	"3rdparty":    true,
	"deploy":      true,
}

// NewValidateHeaders creates a new struct used to validate a project.
func NewValidateHeaders(fileNames *[]string, rootDir string, boilerplateDir string, forceExtension string) ValidateHeaders {
	return ValidateHeaders{
		fileNames:      fileNames,
		rootDir:        rootDir,
		boilerplateDir: boilerplateDir,
		forceExtension: forceExtension,
	}
}

// Validate iterates over a set of files and returns a slice of files that
// do not have the proper headers. A slice of files are returned that containes
// the filenames of files that do not have a header that matches the boilerplate.
func (v ValidateHeaders) Validate() (nonconformingFilesPtr *[]string, err error) {

	// This regular expression is used repace a boilerplate year with the
	// current year.
	// TODO we may want to put a placeholder in the boilerplate and
	// replace that with the current year.
	reg, err := regexp.Compile(DateRegex)
	if err != nil {
		return nil, err
	}

	// Read in all of the boilerplate files
	references, err := v.readGlob("/*.txt", reg)
	if err != nil {
		return nil, err
	}
	if references == nil || len(*references) == 0 {
		return nil, errors.New("unable to find any boilerplates")
	}

	v.references = references

	// Read in all of the preambles
	preambles, err := v.readGlob("/boilerplate.*.preamble", nil)
	if err != nil || len(*preambles) == 0 {
		return nil, err
	}

	if preambles == nil {
		return nil, errors.New("unable to find any preambles")
	}

	v.preambles = preambles

	// Get all the files to validate.
	files, err := v.getFiles()
	if err != nil {
		return nil, err
	}
	if files == nil || len(*files) == 0 {
		return nil, errors.New("no files found")
	}

	var nonconformingFiles []string

	// Check if any file has an invalid header.
	for _, filename := range *files {
		valid, err := v.hasValidHeader(filename)
		if err != nil {
			return nil, err
		}
		if !valid {
			nonconformingFiles = append(nonconformingFiles, filename)
		}
	}

	if len(nonconformingFiles) != 0 {
		sort.Strings(nonconformingFiles)
		return &nonconformingFiles, nil
	}

	return nil, nil
}

func (v ValidateHeaders) readGlob(glob string, regex *regexp.Regexp) (filesPtr *map[string][]string, err error) {

	files, err := filepath.Glob(v.boilerplateDir + glob)
	if err != nil {
		return nil, err
	}

	s := make(map[string][]string)
	for _, f := range files {
		key := path.Base(f)
		key = strings.Split(key, ".")[1]
		content, err := readLines(f, regex)
		if err != nil {
			return nil, err
		}
		s[key] = content
	}

	return &s, err
}

func (v ValidateHeaders) hasValidHeader(filename string) (bool, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return false, err
	}

	data := string(content)
	basename := path.Base(filename)
	extension := removeDot(path.Ext(filename))

	if v.forceExtension != "" {
		extension = v.forceExtension
	} else if extension == "" {
		extension = basename
	}

	m := *v.preambles
	preambleSlice := m[extension]
	if len(preambleSlice) != 0 {
		preamble := regexp.QuoteMeta(strings.Join(preambleSlice, "\n"))
		regex, err := regexp.Compile(fmt.Sprintf("^(%s.*\n)\n*", preamble))
		if err != nil {
			return false, err
		}
		data = regex.ReplaceAllString(data, "")
	}

	m = *v.references
	ref := m[extension]
	dataSlice := strings.Split(data, "\n")

	// if our test file is smaller than the reference it surely fails!
	if len(ref) > len(dataSlice) {
		return false, nil
	}
	// truncate our file to the same number of lines as the reference file
	dataSlice = dataSlice[:len(ref)]

	for i, refLine := range ref {
		// If our test file line does not match the same line as the ref
		// then we fail.
		if dataSlice[i] != refLine {
			return false, nil
		}
	}

	return true, nil
}

func (v ValidateHeaders) getFiles() (filesPtr *[]string, err error) {

	if v.fileNames != nil && len(*v.fileNames) != 0 {
		return v.fileNames, nil
	}

	var files []string

	fileExtensions := make(map[string]bool)
	ext := *v.references
	for k := range ext {
		fileExtensions[k] = true
	}

	err = filepath.Walk(v.rootDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", p, err)
			return err
		}
		if info.IsDir() && SkippedPaths[info.Name()] {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		} else if fileExtensions[removeDot(path.Ext(p))] || fileExtensions[path.Base(p)] {
			files = append(files, p)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", v.rootDir, err)
		return nil, err
	}

	return &files, nil
}

func removeDot(str string) string {
	if len(str) == 0 {
		return str
	}
	first := str[0:1]
	if first == "." {
		_, i := utf8.DecodeRuneInString(str)
		return str[i:]
	}
	return str
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string, reg *regexp.Regexp) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	year, _, _ := time.Now().Date()
	for scanner.Scan() {
		text := scanner.Text()
		if reg != nil {
			text = reg.ReplaceAllString(text, strconv.Itoa(year))
		}
		lines = append(lines, text)
	}
	return lines, scanner.Err()
}
