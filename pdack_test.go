package main

import (
	"os"
	"testing"
)

var TestFiles = []struct {
	filename string
	passing  bool
}{
	{"/pdack_sample.conf", true},
	{"/_example/invalid_filename.conf", false},
	{"/_example/empty.conf", false},
	{"/_example/missing_email.conf", false},
}

// TestReadConfigFile tests the configuation file is being read correctly
func TestReadConfigFile(t *testing.T) {
	var conf PagerDutyConfig
	pwd, _ := os.Getwd()
	for _, testFile := range TestFiles {
		res, _ := readConfigFile(pwd+testFile.filename, &conf)
		if res != testFile.passing {
			if testFile.passing {
				t.Errorf("Expected %s to pass, but it did not", testFile.filename)
			} else {
				t.Errorf("Expected %s to fail, but it did not", testFile.filename)
			}
		}
	}
}
