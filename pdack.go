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

var filename = flag.String("conf", "pdack_sample.conf", "Configuration file")

func readConfigFile(configFileName string, conf *PagerDutyConfig) (success bool, md toml.MetaData) {
	md, err := toml.DecodeFile(configFileName, *conf)
	if err != nil {
		log.Printf("An error occured while reading the configuation file: %s", err)
		return false, md
	}
	if len(md.Keys()) < 3 {
		log.Print("An error occured while reading the configuation file, required key is missing")
		return false, md
	}
	return true, md
}

func getConfigFile(conf *PagerDutyConfig) (md toml.MetaData) {
	pwd, _ := os.Getwd()
	success, md := readConfigFile(pwd+"/"+*filename, conf)
	if success {
		fmt.Printf("%+v\n", md)
	} else {
		os.Exit(1)
	}
	return md
}

func main() {
	var conf PagerDutyConfig
	flag.Parse()
	getConfigFile(&conf)
}
