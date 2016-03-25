package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// PagerDutyConfig contains all the PagerDuty required information
type PagerDutyConfig struct {
	apiKey  string
	email   string
	account string
}

// PagerDutyConfigKeys contains all the keys of PagerDutyConfig
var PagerDutyConfigKeys = []string{
	"apiKey",
	"email",
	"account",
}

var filename = flag.String("conf", "pdack_sample.conf", "Configuration file")

func readConfigFile(configFileName string, conf *PagerDutyConfig) (success bool, md toml.MetaData) {
	md, err := toml.DecodeFile(configFileName, *conf)
	if err != nil {
		log.Printf("An error occured while reading the configuation file: %s", err)
		return false, md
	}
	if len(md.Keys()) < 3 {
		for _, key := range PagerDutyConfigKeys {
			if !md.IsDefined(key) {
				log.Printf("An error occured while reading the configuation file, %s key is missing", key)
			}
		}
		return false, md
	}
	return true, md
}

func getConfigFile(conf *PagerDutyConfig) (success bool, md toml.MetaData) {
	pwd, _ := os.Getwd()
	flag.Parse()
	return readConfigFile(pwd+"/"+*filename, conf)
}

func main() {
	var conf PagerDutyConfig
	success, md := getConfigFile(&conf)
	if success {
		fmt.Print(md)
	} else {
		os.Exit(1)
	}
}
