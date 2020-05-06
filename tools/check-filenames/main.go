// Copyright 2019-2020 Celer Network
//
// Check that file and directory names within the given directory
// follow the naming guidelines.
//
// No output if there are no issues, otherwise print one line per
// file or directory name that fail the check.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	SKIP_DIRS = ".git"              // comma-separated list
	NO_UPPER  = ".go,.sh,.py,.c,.h" // comma-separated list
)

type StringSet map[string]bool

var (
	skipDirs   StringSet // set of directories to skip
	noUpper    StringSet // set of file suffixes w/o uppercase
	hasInvalid bool      // invalid files/dirs found?
	hasError   bool      // error traversing the tree?
)

// Callback of the walk traversal.
func walkCallback(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Printf("access error %s\n", err)
		hasError = true
		return nil
	}

	valid := false
	ftype := ""
	name := info.Name()

	if info.IsDir() {
		if skipDirs[name] {
			return filepath.SkipDir
		}

		valid = validDir(name)
		ftype = "dir"
	} else {
		valid = validFile(name)
		ftype = "file"
	}

	if !valid {
		fmt.Printf("invalid %s: %s\n", ftype, path)
		hasInvalid = true
	}

	return nil
}

// Directory names cannot use underscore.
func validDir(name string) bool {
	return !strings.Contains(name, "_")
}

// File names cannot use dash.  Also files of certain types
// (by suffix) cannot use uppercase letters.
func validFile(name string) bool {
	if strings.Contains(name, "-") {
		return false
	}

	suffix := filepath.Ext(name)
	if noUpper[suffix] {
		if strings.ToLower(name) != name {
			return false
		}
	}

	return true
}

func makeStringSet(list string) StringSet {
	set := make(StringSet)
	for _, val := range strings.Split(list, ",") {
		val = strings.TrimSpace(val)
		if val != "" {
			set[val] = true
		}
	}

	return set
}

func printUsage() {
	fmt.Println("check_filenames <rootDir>")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		printUsage()
		os.Exit(1)
	}

	root := args[0]
	skipDirs = makeStringSet(SKIP_DIRS)
	noUpper = makeStringSet(NO_UPPER)

	err := filepath.Walk(root, walkCallback)
	if err != nil {
		fmt.Printf("cannot walk tree %s: %s\n", root, err)
		os.Exit(1)
	}

	if hasInvalid || hasError {
		os.Exit(1)
	}
}
