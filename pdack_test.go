package main

import (
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
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
	pwd, _ := os.Getwd()
	var conf PagerDutyConfig
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

// TestgetConfigFile tests the
func TestGetConfigFile(t *testing.T) {
	var conf PagerDutyConfig
	var md toml.MetaData
	returned_md := getConfigFile(&conf)
	assert.NotEqual(t, md, returned_md, "The config from the PagerDutyConfig should not be empty")
}

func TestMain(t *testing.T) {
	main()
}
