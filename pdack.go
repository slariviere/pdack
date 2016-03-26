package main

import (
	"flag"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// PagerDutyConfig contains all the PagerDuty required information
type PagerDutyConfig struct {
	APIKey  string
	Email   string
	Account string
}

// PagerDutyConfigKeys contains all the keys of PagerDutyConfig
var PagerDutyConfigKeys = []string{
	"apiKey",
	"email",
	"account",
}

var filename = flag.String("conf", "pdack_sample.conf", "Configuration file")
var config PagerDutyConfig

func readConfigFile(configFileName string) (md toml.MetaData, success bool) {
	md, err := toml.DecodeFile(configFileName, &config)
	if err != nil {
		log.Printf("An error occured while reading the configuation file: %s", err)
		return md, false
	}
	if len(md.Keys()) < 3 {
		for _, key := range PagerDutyConfigKeys {
			if !md.IsDefined(key) {
				log.Printf("An error occured while reading the configuation file, %s key is missing", key)
			}
		}
		return md, false
	}
	return md, true
}

func getConfigFile() (md toml.MetaData, success bool) {
	pwd, _ := os.Getwd()
	flag.Parse()
	return readConfigFile(pwd + "/" + *filename)
}

func getAssignedPDIncidents() {
}

func main() {
	_, success := getConfigFile()

	if success {
		getAssignedPDIncidents()
	} else {
		os.Exit(1)
	}
}
