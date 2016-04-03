package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"gopkg.in/h2non/gock.v0"

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
	{"/_example/empty.conf", false, []string{"apiKey", "userID", "account"}},
	{"/_example/missing_email.conf", false, []string{"userID"}},
	{"/_example/missing_apiKey.conf", false, []string{"apiKey"}},
	{"/_example/missing_account.conf", false, []string{"account"}},
}

var b bytes.Buffer
var traceBuffer = bufio.NewWriter(&b)
var testExitCode = 0

func init() {
	log.SetOutput(traceBuffer)
	waitDelay = 0
	// Desactivate the os.Exit duing the tests
	myPrivateExitFunction = func(c int) {
		testExitCode = c
	}
}

// TestReadConfigFile tests the configuation file is being read correctly
func TestReadConfigFile(t *testing.T) {
	pwd, _ := os.Getwd()
	for _, testFile := range TestFiles {
		_, res := readConfigFile(pwd + testFile.filename)
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
	var md toml.MetaData
	_, returnedmd := getConfigFile()
	assert.NotEqual(t, md, returnedmd, "The config from the PagerDutyConfig should not be empty")
}

// TestgetConfigFile tests reading the conf argument
func TestGetConfigFile(t *testing.T) {
	var md toml.MetaData
	for _, testFile := range TestFiles {
		os.Args = []string{os.Args[0], "--conf=" + testFile.filename}
		returnedmd, res := getConfigFile()
		if res != testFile.passing {
			if testFile.passing {
				assert.NotEqual(t, md, returnedmd, "The config from the PagerDutyConfig should not be empty")
			} else {
				t.Errorf("Expected %s to fail, but it did not", testFile.filename)
			}
		}
	}
}

func TestGetPDURL(t *testing.T) {
	assert.Equal(t, getPDURL(), "https://"+config.Account+".pagerduty.com", "Invalid url returned by getPDURL")
}

func TestBuidIcindentURL(t *testing.T) {
	assert.Equal(t, buidIcindentURL(), "https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user=XXXXXXX", "Invalid url returned by buidIcindentURL")
}

func TestBuidAcknowledgeURL(t *testing.T) {
	assert.Equal(t, buidAcknowledgeURL("123"), "https://"+config.Account+".pagerduty.com/api/v1/incidents/123/acknowledge", "Invalid url returned by buidAcknowledgeURL")
}

func TestGetAssignedPDIncidents(t *testing.T) {
	defer gock.Off()

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(200).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	assert.Equal(t, getAssignedPDIncidents(), true, "Response code is 200, getAssignedPDIncidents should return true")

	assert.Equal(t, gock.IsDone(), true, "Did not sent the planned request to PD")
}

func TestGetAssignedPDIncidentsBadRequest(t *testing.T) {
	defer gock.Off()

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(400).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	assert.Equal(t, getAssignedPDIncidents(), false, "Response code is 400, getAssignedPDIncidents should return false")

	assert.Equal(t, gock.IsDone(), true, "Did not sent the planned request to PD")
}

func TestGetAssignedPDIncidentsRetriesFails(t *testing.T) {
	pdRetryCount = 0
	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	assert.Equal(t, getAssignedPDIncidents(), false, "Response code is 500 too many times, getAssignedPDIncidents should return false")
	assert.Equal(t, gock.IsDone(), true, "Did not sent the planned request to PD")
}

func TestGetAssignedPDIncidentsRetriesSucceed(t *testing.T) {
	pdRetryCount = 0
	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(200).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	assert.Equal(t, getAssignedPDIncidents(), true, "Response code is 200 at the last moment, getAssignedPDIncidents should return true")
	assert.Equal(t, gock.IsDone(), true, "Did not sent the planned request to PD")
}

func TestGetAssignedPDIncidentsWithAck(t *testing.T) {
	pdRetryCount = 0
	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(200).
		BodyString(`{"incidents":[{"id":"PO7FKW9","incident_number":111661,"created_on":"2016-04-03T01:54:02Z","status":"triggered","pending_actions":[],"html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9","incident_key":"66274f0746384df2ad51c04c2d4069bb","service":{"id":"P7C31P0","name":"TEST_SERVICE","html_url":"https://your_account.pagerduty.com/services/P7C31P0","deleted_at":null,"description":""},"escalation_policy":{"id":"P5W7JL2","name":"MO - Sebastien Lariviere","deleted_at":null},"assigned_to_user":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT"},"trigger_summary_data":{"subject":"t"},"trigger_details_html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF","trigger_type":"web_trigger","last_status_change_on":"2016-04-03T01:54:02Z","last_status_change_by":null,"number_of_escalations":0,"assigned_to":[{"at":"2016-04-03T01:54:02Z","object":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT","type":"user"}}],"urgency":"low"}],"limit":100,"offset":0,"total":1}`)

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents/PO7FKW9/acknowledge").
		MatchHeader("Authorization", config.APIKey).
		Reply(200).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	assert.Equal(t, getAssignedPDIncidents(), true, "Should send ack to the mentionned icident ID")
	assert.Equal(t, gock.IsDone(), true, "Did not sent the planned request to PD")
}

func TestGetAssignedPDIncidentsWithAcks(t *testing.T) {
	pdRetryCount = 0
	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(200).
		BodyString(`{ "incidents": [ { "id": "PO7FKW9", "incident_number": 111661, "created_on": "2016-04-03T01:54:02Z", "status": "triggered", "pending_actions": [], "html_url": "https://your_account.pagerduty.com/incidents/PO7FKW9", "incident_key": "66274f0746384df2ad51c04c2d4069bb", "service": { "id": "P7C31P0", "name": "TEST_SERVICE", "html_url": "https://your_account.pagerduty.com/services/P7C31P0", "deleted_at": null, "description": "" }, "escalation_policy": { "id": "P5W7JL2", "name": "MO - Sebastien Lariviere", "deleted_at": null }, "assigned_to_user": { "id": "PJGAQGT", "name": "Sébastien Larivière", "email": "sebastien@lariviere.me", "html_url": "https://your_account.pagerduty.com/users/PJGAQGT" }, "trigger_summary_data": { "subject": "t" }, "trigger_details_html_url": "https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF", "trigger_type": "web_trigger", "last_status_change_on": "2016-04-03T01:54:02Z", "last_status_change_by": null, "number_of_escalations": 0, "assigned_to": [ { "at": "2016-04-03T01:54:02Z", "object": { "id": "PJGAQGT", "name": "Sébastien Larivière", "email": "sebastien@lariviere.me", "html_url": "https://your_account.pagerduty.com/users/PJGAQGT", "type": "user" } } ], "urgency": "low" }, { "id": "PO7FKW0", "incident_number": 111661, "created_on": "2016-04-03T01:54:02Z", "status": "triggered", "pending_actions": [], "html_url": "https://your_account.pagerduty.com/incidents/PO7FKW0", "incident_key": "66274f0746384df2ad51c04c2d4069ba", "service": { "id": "P7C31P0", "name": "TEST_SERVICE", "html_url": "https://your_account.pagerduty.com/services/P7C31P0", "deleted_at": null, "description": "" }, "escalation_policy": { "id": "P5W7JL2", "name": "MO - Sebastien Lariviere", "deleted_at": null }, "assigned_to_user": { "id": "PJGAQGT", "name": "Sébastien Larivière", "email": "sebastien@lariviere.me", "html_url": "https://your_account.pagerduty.com/users/PJGAQGT" }, "trigger_summary_data": { "subject": "t2" }, "trigger_details_html_url": "https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF", "trigger_type": "web_trigger", "last_status_change_on": "2016-04-03T01:54:02Z", "last_status_change_by": null, "number_of_escalations": 0, "assigned_to": [ { "at": "2016-04-03T01:54:02Z", "object": { "id": "PJGAQGT", "name": "Sébastien Larivière", "email": "sebastien@lariviere.me", "html_url": "https://your_account.pagerduty.com/users/PJGAQGT", "type": "user" } } ], "urgency": "low" } ], "limit": 100, "offset": 0, "total": 1 }`)

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents/PO7FKW9/acknowledge").
		MatchHeader("Authorization", config.APIKey).
		Reply(200).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents/PO7FKW0/acknowledge").
		MatchHeader("Authorization", config.APIKey).
		Reply(200).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	assert.Equal(t, getAssignedPDIncidents(), true, "Should send ack to the mentionned icident ID")
	assert.Equal(t, gock.IsDone(), true, "Did not sent the planned request to PD")
}

func TestGetAssignedPDIncidentsWithIcidentsAcked(t *testing.T) {
	pdRetryCount = 0
	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(200).
		BodyString(`{ "incidents": [ { "id": "PO7FKW9", "incident_number": 111661, "created_on": "2016-04-03T01:54:02Z", "status": "acknowledged", "pending_actions": [], "html_url": "https://your_account.pagerduty.com/incidents/PO7FKW9", "incident_key": "66274f0746384df2ad51c04c2d4069bb", "service": { "id": "P7C31P0", "name": "TEST_SERVICE", "html_url": "https://your_account.pagerduty.com/services/P7C31P0", "deleted_at": null, "description": "" }, "escalation_policy": { "id": "P5W7JL2", "name": "MO - Sebastien Lariviere", "deleted_at": null }, "assigned_to_user": { "id": "PJGAQGT", "name": "Sébastien Larivière", "email": "sebastien@lariviere.me", "html_url": "https://your_account.pagerduty.com/users/PJGAQGT" }, "trigger_summary_data": { "subject": "t" }, "trigger_details_html_url": "https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF", "trigger_type": "web_trigger", "last_status_change_on": "2016-04-03T01:54:02Z", "last_status_change_by": null, "number_of_escalations": 0, "assigned_to": [ { "at": "2016-04-03T01:54:02Z", "object": { "id": "PJGAQGT", "name": "Sébastien Larivière", "email": "sebastien@lariviere.me", "html_url": "https://your_account.pagerduty.com/users/PJGAQGT", "type": "user" } } ], "urgency": "low" }, { "id": "PO7FKW0", "incident_number": 111661, "created_on": "2016-04-03T01:54:02Z", "status": "acknowledged", "pending_actions": [], "html_url": "https://your_account.pagerduty.com/incidents/PO7FKW0", "incident_key": "66274f0746384df2ad51c04c2d4069ba", "service": { "id": "P7C31P0", "name": "TEST_SERVICE", "html_url": "https://your_account.pagerduty.com/services/P7C31P0", "deleted_at": null, "description": "" }, "escalation_policy": { "id": "P5W7JL2", "name": "MO - Sebastien Lariviere", "deleted_at": null }, "assigned_to_user": { "id": "PJGAQGT", "name": "Sébastien Larivière", "email": "sebastien@lariviere.me", "html_url": "https://your_account.pagerduty.com/users/PJGAQGT" }, "trigger_summary_data": { "subject": "t2" }, "trigger_details_html_url": "https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF", "trigger_type": "web_trigger", "last_status_change_on": "2016-04-03T01:54:02Z", "last_status_change_by": null, "number_of_escalations": 0, "assigned_to": [ { "at": "2016-04-03T01:54:02Z", "object": { "id": "PJGAQGT", "name": "Sébastien Larivière", "email": "sebastien@lariviere.me", "html_url": "https://your_account.pagerduty.com/users/PJGAQGT", "type": "user" } } ], "urgency": "low" } ], "limit": 100, "offset": 0, "total": 1 }`)

	assert.Equal(t, getAssignedPDIncidents(), true, "Should send ack to the mentionned icident ID")
	assert.Equal(t, gock.IsDone(), true, "Did not sent the planned request to PD")
}

func TestGetAssignedPDIncidentsWithAckBadRequest(t *testing.T) {
	pdRetryCount = 0
	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents?assigned_to_user="+config.UserID).
		MatchHeader("Authorization", config.APIKey).
		Reply(200).
		BodyString(`{"incidents":[{"id":"PO7FKW9","incident_number":111661,"created_on":"2016-04-03T01:54:02Z","status":"triggered","pending_actions":[],"html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9","incident_key":"66274f0746384df2ad51c04c2d4069bb","service":{"id":"P7C31P0","name":"TEST_SERVICE","html_url":"https://your_account.pagerduty.com/services/P7C31P0","deleted_at":null,"description":""},"escalation_policy":{"id":"P5W7JL2","name":"MO - Sebastien Lariviere","deleted_at":null},"assigned_to_user":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT"},"trigger_summary_data":{"subject":"t"},"trigger_details_html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF","trigger_type":"web_trigger","last_status_change_on":"2016-04-03T01:54:02Z","last_status_change_by":null,"number_of_escalations":0,"assigned_to":[{"at":"2016-04-03T01:54:02Z","object":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT","type":"user"}}],"urgency":"low"}],"limit":100,"offset":0,"total":1}`)

	gock.New("https://"+config.Account+".pagerduty.com/api/v1/incidents/PO7FKW9/acknowledge").
		MatchHeader("Authorization", config.APIKey).
		Reply(400).
		BodyString(`{"incidents":[],"limit":100,"offset":0,"total":0}`)

	assert.Equal(t, getAssignedPDIncidents(), false, "Should send ack to the mentionned icident ID, then get a 400, thus failing")
	assert.Equal(t, gock.IsDone(), true, "Did not sent the planned request to PD")
}

// TestMain tests the main function
func TestMain(t *testing.T) {
	os.Args = []string{os.Args[0], "--conf=pdack_sample.conf"}

	pdRetryCount = 0
	testExitCode = 0
	gock.New("https://your_account.pagerduty.com/api/v1/incidents?assigned_to_user=XXXXXXX").
		Reply(200).
		BodyString(`{"incidents":[{"id":"PO7FKW9","incident_number":111661,"created_on":"2016-04-03T01:54:02Z","status":"triggered","pending_actions":[],"html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9","incident_key":"66274f0746384df2ad51c04c2d4069bb","service":{"id":"P7C31P0","name":"TEST_SERVICE","html_url":"https://your_account.pagerduty.com/services/P7C31P0","deleted_at":null,"description":""},"escalation_policy":{"id":"P5W7JL2","name":"MO - Sebastien Lariviere","deleted_at":null},"assigned_to_user":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT"},"trigger_summary_data":{"subject":"t"},"trigger_details_html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF","trigger_type":"web_trigger","last_status_change_on":"2016-04-03T01:54:02Z","last_status_change_by":null,"number_of_escalations":0,"assigned_to":[{"at":"2016-04-03T01:54:02Z","object":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT","type":"user"}}],"urgency":"low"}],"limit":100,"offset":0,"total":1}`)

	// Faking PD outage so the program exit
	gock.New("https://your_account.pagerduty.com/api/v1/incidents/PO7FKW9/acknowledge").
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[{"id":"PO7FKW9","incident_number":111661,"created_on":"2016-04-03T01:54:02Z","status":"triggered","pending_actions":[],"html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9","incident_key":"66274f0746384df2ad51c04c2d4069bb","service":{"id":"P7C31P0","name":"TEST_SERVICE","html_url":"https://your_account.pagerduty.com/services/P7C31P0","deleted_at":null,"description":""},"escalation_policy":{"id":"P5W7JL2","name":"MO - Sebastien Lariviere","deleted_at":null},"assigned_to_user":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT"},"trigger_summary_data":{"subject":"t"},"trigger_details_html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF","trigger_type":"web_trigger","last_status_change_on":"2016-04-03T01:54:02Z","last_status_change_by":null,"number_of_escalations":0,"assigned_to":[{"at":"2016-04-03T01:54:02Z","object":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT","type":"user"}}],"urgency":"low"}],"limit":100,"offset":0,"total":1}`)

	gock.New("https://your_account.pagerduty.com/api/v1/incidents/PO7FKW9/acknowledge").
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[{"id":"PO7FKW9","incident_number":111661,"created_on":"2016-04-03T01:54:02Z","status":"triggered","pending_actions":[],"html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9","incident_key":"66274f0746384df2ad51c04c2d4069bb","service":{"id":"P7C31P0","name":"TEST_SERVICE","html_url":"https://your_account.pagerduty.com/services/P7C31P0","deleted_at":null,"description":""},"escalation_policy":{"id":"P5W7JL2","name":"MO - Sebastien Lariviere","deleted_at":null},"assigned_to_user":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT"},"trigger_summary_data":{"subject":"t"},"trigger_details_html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF","trigger_type":"web_trigger","last_status_change_on":"2016-04-03T01:54:02Z","last_status_change_by":null,"number_of_escalations":0,"assigned_to":[{"at":"2016-04-03T01:54:02Z","object":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT","type":"user"}}],"urgency":"low"}],"limit":100,"offset":0,"total":1}`)

	gock.New("https://your_account.pagerduty.com/api/v1/incidents/PO7FKW9/acknowledge").
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[{"id":"PO7FKW9","incident_number":111661,"created_on":"2016-04-03T01:54:02Z","status":"triggered","pending_actions":[],"html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9","incident_key":"66274f0746384df2ad51c04c2d4069bb","service":{"id":"P7C31P0","name":"TEST_SERVICE","html_url":"https://your_account.pagerduty.com/services/P7C31P0","deleted_at":null,"description":""},"escalation_policy":{"id":"P5W7JL2","name":"MO - Sebastien Lariviere","deleted_at":null},"assigned_to_user":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT"},"trigger_summary_data":{"subject":"t"},"trigger_details_html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF","trigger_type":"web_trigger","last_status_change_on":"2016-04-03T01:54:02Z","last_status_change_by":null,"number_of_escalations":0,"assigned_to":[{"at":"2016-04-03T01:54:02Z","object":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT","type":"user"}}],"urgency":"low"}],"limit":100,"offset":0,"total":1}`)

	gock.New("https://your_account.pagerduty.com/api/v1/incidents/PO7FKW9/acknowledge").
		MatchHeader("Authorization", config.APIKey).
		Reply(500).
		BodyString(`{"incidents":[{"id":"PO7FKW9","incident_number":111661,"created_on":"2016-04-03T01:54:02Z","status":"triggered","pending_actions":[],"html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9","incident_key":"66274f0746384df2ad51c04c2d4069bb","service":{"id":"P7C31P0","name":"TEST_SERVICE","html_url":"https://your_account.pagerduty.com/services/P7C31P0","deleted_at":null,"description":""},"escalation_policy":{"id":"P5W7JL2","name":"MO - Sebastien Lariviere","deleted_at":null},"assigned_to_user":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT"},"trigger_summary_data":{"subject":"t"},"trigger_details_html_url":"https://your_account.pagerduty.com/incidents/PO7FKW9/log_entries/Q0P5VKOXNK4MSF","trigger_type":"web_trigger","last_status_change_on":"2016-04-03T01:54:02Z","last_status_change_by":null,"number_of_escalations":0,"assigned_to":[{"at":"2016-04-03T01:54:02Z","object":{"id":"PJGAQGT","name":"S\u00e9bastien Larivi\u00e8re","email":"sebastien@lariviere.me","html_url":"https://your_account.pagerduty.com/users/PJGAQGT","type":"user"}}],"urgency":"low"}],"limit":100,"offset":0,"total":1}`)

	main()
	assert.Equal(t, gock.IsDone(), true, "Did not sent the planned request to PD")
	assert.Equal(t, testExitCode, 1, "Program should have exited with code 1")
}

// TestMain tests the main function
func TestMainBadConfigurationFile(t *testing.T) {
	os.Args = []string{os.Args[0], "--conf=pdack_does_not_exists.conf"}
	pdRetryCount = 0
	testExitCode = 0
	main()
	assert.Equal(t, testExitCode, 1, "Program should have exited with code 1")
}
