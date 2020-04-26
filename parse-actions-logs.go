// Parse the logs from GitHub Actions
package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Flags
	outputDir = flag.String("o", "output", "Output directory")
	// Globals
	matchProject = regexp.MustCompile(`^([\w-]+)/([\w-]+)$`)
	project      string
	baseURL      = "https://api.github.com/repos/"
	errors       int
	version      = "development version" // overridden by goreleaser
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [options] <log.zip>+
Version: %s

Parse the logs fetched from GitHub actions. These logs should be the
zip files downloaded for a complete workflow run.

They can be downloaded by https://github.com/ncw/fetch-actions-logs or
by downloading from the web UI.

Example usage:

parse-actions-logs logs.zip  logs2.zip

Full options:
`, os.Args[0], version)
	flag.PrintDefaults()
}

// All the log lines come with a timestamp prefix like this
// 2020-04-15T12:19:01.5953417Z
var timestampRe = regexp.MustCompile(`(?m)(^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d*)?Z )`)

// Search for test failure markers
var failRe = regexp.MustCompile(`(?m)^\s*--- FAIL: (Test.*?) \(`)

// Look for individual test results
// ok         github.com/rclone/rclone/lib/mmap       0.130s
// ?          github.com/rclone/rclone/lib/oauthutil  [no test files]
// FAIL       github.com/rclone/rclone/lib/pacer      0.813s
var splitRe = regexp.MustCompile(`(?m)^(ok  |FAIL|\?   )\t([^\t]+)\t.*$`)

// Munge a string into something acceptable as a filename
//
/// FIXME this probably needs a bit more work on Windows
func toFileName(s string) string {
	s = strings.Replace(s, "/", "／", -1)
	s = strings.Replace(s, "\\", "＼", -1)
	return s
}

// findFailures looks for all the tests which failed and returns the
// deepest failure excluding the parent
func findFailures(buf []byte) []string {
	var (
		failedTests    []string
		excludeParents = map[string]struct{}{}
	)
	for _, matches := range failRe.FindAllSubmatch(buf, -1) {
		failedTest := string(matches[1])
		failedTests = append(failedTests, failedTest)
		// Find all the parents of this test
		parts := strings.Split(failedTest, "/")
		for i := len(parts) - 1; i >= 1; i-- {
			excludeParents[strings.Join(parts[:i], "/")] = struct{}{}
		}
	}
	// Exclude the parents
	var newTests = failedTests[:0]
	for _, failedTest := range failedTests {
		if _, excluded := excludeParents[failedTest]; !excluded {
			newTests = append(newTests, failedTest)
		}
	}
	failedTests = newTests
	return failedTests
}

// This is called to save a failure
//
// We just use the first test failure as the failure name
func saveFailure(zipFile, name, module, failedTest string, buf []byte) {
	dir := filepath.Join(*outputDir, toFileName(module), toFileName(failedTest))
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		log.Printf("Failed to create output dir: %v", err)
		return
	}
	zipName := filepath.Base(zipFile)
	zipName = zipName[:len(zipName)-len(filepath.Ext(zipName))]
	fileName := filepath.Join(dir, toFileName(name+"-"+zipName+".txt"))
	err = ioutil.WriteFile(fileName, buf, 0666)
	if err != nil {
		log.Printf("Failed to write output file: %v", err)
		return
	}
	log.Printf("saved failure in %q", fileName)
}

func parseFile(zipFile, name string, buf []byte) error {
	// remove \r from the file
	buf = bytes.Replace(buf, []byte{'\r'}, []byte{}, -1)
	// remove timestamps from start of lines
	buf = timestampRe.ReplaceAll(buf, []byte{})

	parts := splitRe.FindAllSubmatchIndex(buf, -1)
	if len(parts) == 0 {
		// log.Printf("%s: no go tests", name)
		return nil // no go tests here
	}
	prev := 0
	for _, part := range parts {
		status := string(bytes.TrimSpace(buf[part[2]:part[3]]))
		module := string(buf[part[4]:part[5]])
		// log.Printf("%q: %q", module, status)
		end := part[1]
		if status == "FAIL" {
			toSearch := buf[prev:end]
			// log.Printf("-----------\n%s\n------------", toSearch)
			failedTests := findFailures(toSearch)
			if len(failedTests) != 0 {
				// log.Printf("%s: %s: FailedTests %v", name, module, failedTests)
				saveFailure(zipFile, name, module, failedTests[0], toSearch)
			}
		}
		prev = end
	}

	return nil
}

func parseZip(zipFile string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		// We only need to look at files in subdirectories as
		// the files in the top level contain the same info
		if !strings.Contains(f.Name, "/") {
			continue
		}
		// log.Printf("%s: Reading %s", zipFile, f.Name)
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open zipped file: %w", err)
		}
		buf, err := ioutil.ReadAll(rc)
		if err != nil {
			return fmt.Errorf("failed to read zipped file: %w", err)
		}
		rc.Close()
		err = parseFile(zipFile, f.Name, buf)
		if err != nil {
			log.Printf("failed to parse zipped file: %v", err)
		}
	}
	return nil
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
		log.Fatal("Please supply some zip files with logs in")
	}
	zipFiles := args
	err := os.MkdirAll(*outputDir, 0777)
	if err != nil {
		log.Fatalf("Failed to create output dir: %v", err)
	}
	for _, zipFile := range zipFiles {
		err := parseZip(zipFile)
		if err != nil {
			log.Printf("Failed to parse %q: %v", zipFile, err)
		}
	}
}
