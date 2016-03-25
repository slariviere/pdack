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

// TestgetConfigFile tests reading the default value of GetConfigFile
func TestGetConfigFileDefault(t *testing.T) {
	var conf PagerDutyConfig
	var md toml.MetaData
	_, returnedmd := getConfigFile(&conf)
	assert.NotEqual(t, md, returnedmd, "The config from the PagerDutyConfig should not be empty")
}

// TestgetConfigFile tests reading the conf argument
func TestGetConfigFile(t *testing.T) {
	var conf PagerDutyConfig
	var md toml.MetaData
	for _, testFile := range TestFiles {
		os.Args = []string{os.Args[0], "--conf=" + testFile.filename}
		res, returnedmd := getConfigFile(&conf)
		if res != testFile.passing {
			if testFile.passing {
				assert.NotEqual(t, md, returnedmd, "The config from the PagerDutyConfig should not be empty")
			} else {
				t.Errorf("Expected %s to fail, but it did not", testFile.filename)
			}
		}
	}
}

// TestMain tests the main function
func TestMain(t *testing.T) {
	os.Args = []string{os.Args[0], "--conf=pdack_sample.conf"}
	main()
}
