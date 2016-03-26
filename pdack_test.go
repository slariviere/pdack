package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

var TestFiles = []struct {
	filename    string
	passing     bool
	errorString []string
}{
	{"/pdack_sample.conf", true, []string{""}},
	{"/_example/invalid_filename.conf", false, []string{"_example/invalid_filename.conf"}},
	{"/_example/empty.conf", false, []string{"apiKey", "email", "account"}},
	{"/_example/missing_email.conf", false, []string{"email"}},
	{"/_example/missing_apiKey.conf", false, []string{"apiKey"}},
	{"/_example/missing_account.conf", false, []string{"account"}},
}

var b bytes.Buffer
var traceBuffer = bufio.NewWriter(&b)

func init() {
	log.SetOutput(traceBuffer)
}

// TestReadConfigFile tests the configuation file is being read correctly
func TestReadConfigFile(t *testing.T) {
	pwd, _ := os.Getwd()
	var conf PagerDutyConfig
	for _, testFile := range TestFiles {
		res, _ := readConfigFile(pwd+testFile.filename, &conf)
		b.Reset()
		traceBuffer.Flush()
		// If the result is not the expected result of the test
		if res != testFile.passing {
			if testFile.passing {
				t.Errorf("Expected %s to pass, but it did not", testFile.filename)
			} else {
				t.Errorf("Expected %s to fail, but it did not", testFile.filename)
			}
		}
		// Check if error message contains intended string(s)
		if !testFile.passing {
			for _, errorString := range testFile.errorString {
				if !strings.Contains(b.String(), errorString) {
					t.Errorf("Expected %s in the error message, but it's missing: %s", errorString, b.String())
				}
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
