package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// PagerDutyConfig contains all the PagerDuty required information
type PagerDutyConfig struct {
	APIKey       string
	UserID       string
	Account      string
	RefreshDelay int
}

// PagerDutyConfigKeys contains all the keys of PagerDutyConfig
var PagerDutyConfigKeys = []string{
	"apiKey",
	"userID",
	"account",
	"refreshDelay",
}

var filename = flag.String("conf", "pdack.conf", "Configuration file")
var maxPDretries = 3
var pdRetryCount = 0
var waitDelay = 1
var config PagerDutyConfig

var myPrivateExitFunction = os.Exit

// Incident type
type Incident struct {
	Incidents []struct {
		Acknowledgers []struct {
			At     string `json:"at"`
			Object struct {
				Email   string `json:"email"`
				HTMLURL string `json:"html_url"`
				ID      string `json:"id"`
				Name    string `json:"name"`
				Type    string `json:"type"`
			} `json:"object"`
		} `json:"acknowledgers"`
		AssignedTo []struct {
			At     string `json:"at"`
			Object struct {
				Email   string `json:"email"`
				HTMLURL string `json:"html_url"`
				ID      string `json:"id"`
				Name    string `json:"name"`
				Type    string `json:"type"`
			} `json:"object"`
		} `json:"assigned_to"`
		AssignedToUser struct {
			Email   string `json:"email"`
			HTMLURL string `json:"html_url"`
			ID      string `json:"id"`
			Name    string `json:"name"`
		} `json:"assigned_to_user"`
		CreatedOn        string `json:"created_on"`
		EscalationPolicy struct {
			DeletedAt interface{} `json:"deleted_at"`
			ID        string      `json:"id"`
			Name      string      `json:"name"`
		} `json:"escalation_policy"`
		HTMLURL            string `json:"html_url"`
		ID                 string `json:"id"`
		IncidentKey        string `json:"incident_key"`
		IncidentNumber     int    `json:"incident_number"`
		LastStatusChangeBy struct {
			Email   string `json:"email"`
			HTMLURL string `json:"html_url"`
			ID      string `json:"id"`
			Name    string `json:"name"`
		} `json:"last_status_change_by"`
		LastStatusChangeOn  string        `json:"last_status_change_on"`
		NumberOfEscalations int           `json:"number_of_escalations"`
		PendingActions      []interface{} `json:"pending_actions"`
		Service             struct {
			DeletedAt   interface{} `json:"deleted_at"`
			Description string      `json:"description"`
			HTMLURL     string      `json:"html_url"`
			ID          string      `json:"id"`
			Name        string      `json:"name"`
		} `json:"service"`
		Status                string `json:"status"`
		TriggerDetailsHTMLURL string `json:"trigger_details_html_url"`
		TriggerSummaryData    struct {
			Subject string `json:"subject"`
		} `json:"trigger_summary_data"`
		TriggerType string `json:"trigger_type"`
		Urgency     string `json:"urgency"`
	} `json:"incidents"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

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

func buildURL(resource string, data url.Values) (builtURL string) {
	u, _ := url.ParseRequestURI(getPDURL())
	u.Path = resource
	u.RawQuery = data.Encode()
	return fmt.Sprintf("%v", u)
}

func buidAcknowledgeURL(id string) (incidentURL string) {
	return buildURL("/api/v1/incidents/"+id+"/acknowledge", url.Values{})
}

func buidIcindentURL() (incidentURL string) {
	data := url.Values{}
	data.Add("assigned_to_user", config.UserID)
	return buildURL("/api/v1/incidents", data)
}

func acknowledgeIncicent(id string) (success bool) {
	urlStr := buidAcknowledgeURL(id)
	req, err := http.NewRequest("PUT", urlStr, strings.NewReader(`{"requester_id": "`+config.UserID+`"}`))
	req.Header.Set("Content-type", "application/json")
	req.Header.Add("Authorization", "Token token="+config.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return true
	}
	if resp.StatusCode == 408 || resp.StatusCode == 500 {
		// There was a recoverable error, retrying in $waitDelay second
		time.Sleep(time.Duration(waitDelay) * time.Second)
		if pdRetryCount < maxPDretries {
			pdRetryCount++
			return acknowledgeIncicent(id)
		}
		return false
	}
	return false
}

func getAssignedPDIncidents() (success bool) {

	nbTriggered := 0
	nbAcknowledged := 0
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

	body, _ := ioutil.ReadAll(resp.Body)
	var incidentResponse Incident
	json.Unmarshal(body, &incidentResponse)
	log.Printf("%d incident found", incidentResponse.Total)
	for _, curentIncident := range incidentResponse.Incidents {
		if curentIncident.Status == "triggered" {
			nbTriggered++
			if acknowledgeIncicent(curentIncident.ID) {
				log.Printf("Incident %s (%s) has been Acknowledged\n", curentIncident.TriggerSummaryData.Subject, curentIncident.ID)
			} else {
				return false
			}
		} else if curentIncident.Status == "acknowledged" {
			nbAcknowledged++
		}
	}
	log.Printf("%d acknowledged, %d triggered", nbAcknowledged, nbTriggered)
	if resp.StatusCode == 200 {
		pdRetryCount = 0
		return true
	}
	if resp.StatusCode == 408 || resp.StatusCode == 500 {
		// There was a recoverable error, retrying in $waitDelay second
		time.Sleep(time.Duration(waitDelay) * time.Second)
		if pdRetryCount < maxPDretries {
			pdRetryCount++
			return getAssignedPDIncidents()
		}
		return false
	}
	return false
}

func main() {
	_, success := getConfigFile()

	if success {
		for true {
			if !getAssignedPDIncidents() {
				myPrivateExitFunction(1)
				return
			}
			time.Sleep(time.Duration(config.RefreshDelay) * time.Second)
		}
	} else {
		myPrivateExitFunction(1)
		return
	}
}
