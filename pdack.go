package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

func getPDURL() (url string) {
	return "https://" + config.Account + ".pagerduty.com"
}

func buidIcindentURL() (incidentURL string) {
	resource := "/api/v1/incidents"
	data := url.Values{}

	u, _ := url.ParseRequestURI(getPDURL())
	u.Path = resource
	u.RawQuery = data.Encode()
	urlStr := fmt.Sprintf("%v", u)
	return urlStr
}

func getAssignedPDIncidents() (success bool) {
	urlStr := buidIcindentURL()
	req, err := http.NewRequest("GET", urlStr, nil)
	req.Header.Set("Content-type", "application/json")
	req.Header.Add("Authorization", "Token token="+config.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func main() {
	_, success := getConfigFile()

	if success {
		getAssignedPDIncidents()
	} else {
		os.Exit(1)
	}
}
