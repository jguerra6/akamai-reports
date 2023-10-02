package report

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/edgegrid"
	"github.com/jguerra6/akamai-reports/command"
	"github.com/jguerra6/akamai-reports/cpcode"
	"github.com/jguerra6/akamai-reports/fileHandling"
	"github.com/jguerra6/akamai-reports/request"
)

const (
	dataBasePath      = "./data"
	cpCodeMapFileName = "cpCode_list.json"
	reportBaseUrl     = "/reporting-api/v1/reports/{{reportName}}/versions/{{reportVersion}}/report-data"
)

var flags = map[string]command.Flag{
	"start": {
		Name:     "start",
		Value:    "",
		Usage:    "Set the start date for your report: YYYY-MM-DD",
		Required: true,
	},
	"end": {
		Name:     "end",
		Value:    "",
		Usage:    "Set the end date for your report: YYYY-MM-DD",
		Required: true,
	},
	//"delimiter": {
	//	Name:     "delimiter",
	//	Value:    ",",
	//	Usage:    "Set the delimiter for your csv report",
	//	Required: false,
	//},
}

type Report struct {
	*command.Command
	edgerc *edgegrid.Config
	flags  map[string]*string
}

func New(edgerc *edgegrid.Config) *Report {

	return &Report{
		Command: command.NewCommand("report", flags),
		edgerc:  edgerc,
	}
}

func loadCPCodesMap() (map[string][]cpcode.CPCode, error) {
	fileName := dataBasePath + "/master/" + cpCodeMapFileName

	byteValue, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var cpCodeMap map[string][]cpcode.CPCode

	err = json.Unmarshal(byteValue, &cpCodeMap)
	if err != nil {
		return nil, err
	}
	return cpCodeMap, nil

}

type Root struct {
	Data []BytesDeliveredReport `json:"data"`
}

type BytesDeliveredReport struct {
	CpCode        string `json:"cpcode" csv:"cpcode"`
	BytesOffload  string `json:"bytesOffload" csv:"bytesOffload"`
	EdgeBytes     string `json:"edgeBytes" csv:"edgeBytes"`
	MidgressBytes string `json:"midgressBytes" csv:"midgressBytes"`
	OriginBytes   string `json:"originBytes" csv:"originBytes"`
	StartDate     string `json:"startDate" csv:"startDate"`
	EndDate       string `json:"endDate" csv:"endDate"`
}

func parseResponse(resp []byte, options map[string]string) ([]BytesDeliveredReport, error) {
	var data Root

	err := json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}

	var bytesDeliveredReport []BytesDeliveredReport

	for _, report := range data.Data {
		report.StartDate = options["start"]
		report.EndDate = options["end"]
		bytesDeliveredReport = append(bytesDeliveredReport, report)
	}

	return bytesDeliveredReport, nil

}

func copyOptionsMap(baseOptions map[string]string) map[string]string {
	newOptions := make(map[string]string)

	for key, value := range baseOptions {
		// Add the modified key-value pair to the modifiedMap
		newOptions[key] = value
	}

	return newOptions

}

// TODO: Add multi-threading
// check for optimizations and breaking into smaller functions
// improve error logging
func (r *Report) getBytesDelivered(accountSwitchKey string, cpCodes *[]cpcode.CPCode, options map[string]string) ([]BytesDeliveredReport, error) {

	reportConfig := map[string]string{
		"reportName":    "bytes-by-cpcode",
		"reportVersion": "1",
	}

	reportOptions := copyOptionsMap(options)
	reportOptions["accountSwitchKey"] = accountSwitchKey

	requestUrl := &url.URL{
		Path: request.ReplaceParams(reportConfig, reportBaseUrl),
	}

	var cpCodeIds []string

	if accountSwitchKey == "B-3-QCCVOP:1-9OGH" {
		fmt.Println("TEST")
	}

	for _, cpCode := range *cpCodes {
		cpCodeId := strconv.Itoa(cpCode.CPCodeId)
		cpCodeIds = append(cpCodeIds, cpCodeId)
	}

	reportOptions["objectIds"] = strings.Join(cpCodeIds[:], ",")

	startDate, err := time.Parse(time.RFC3339, reportOptions["start"])
	if err != nil {
		return nil, err
	}
	endDate, err := time.Parse(time.RFC3339, reportOptions["end"])
	if err != nil {
		return nil, err
	}

	var bytesDeliveredReport []BytesDeliveredReport

	var errors []string

	// Loop through the dates and increase startDate by 1 day until it reaches endDate
	for startDate.Before(endDate) {
		reportOptions["start"] = startDate.Format("2006-01-02T15:04:05Z")
		startDate = startDate.AddDate(0, 0, 1) // Increment startDate by 1 day
		reportOptions["end"] = startDate.Format("2006-01-02T15:04:05Z")

		resp, err := request.AkamaiRequest(r.edgerc, requestUrl.RequestURI(), reportOptions)
		if err != nil {
			//errMsg := fmt.Sprintf("There was an error with the report date: %s. Error: %v", reportOptions["start"], err)
			errors = append(errors, "errMsg")
			continue
		}

		tmp, err := parseResponse(resp, reportOptions)
		if err != nil {
			return nil, err
		}

		bytesDeliveredReport = append(bytesDeliveredReport, tmp...)

	}

	//for _, s := range errors {
	//log.Printf("Error with key: %s, msg: %s", accountSwitchKey, s)
	//}
	startDate, err = time.Parse(time.RFC3339, options["start"])
	duration := endDate.Sub(startDate)

	// Calculate error rate for logging purposes
	days := duration.Hours() / 24.0
	daysInt := int(days)

	errorCount := len(errors)
	codesCount := daysInt
	errRate := errorCount / codesCount
	log.Printf("%s: %d/%d = %d", accountSwitchKey, errorCount, codesCount, errRate)

	return bytesDeliveredReport, nil
}

func (r *Report) Run() error {
	commandFlags := r.Command.Flags()

	//delimiter := []rune(flags["delimiter"].Value)
	delimiter := []rune(",")

	reportOptions := map[string]string{}
	for s, flag := range flags {
		if flag.Required && (commandFlags[s] == nil || *commandFlags[s] == "") {
			return fmt.Errorf(flag.Usage)
		}
		reportOptions[s] = *commandFlags[s]

	}

	reportOptions["start"] += "T00:00:00Z"
	reportOptions["end"] += "T00:00:00Z"

	cpCodeMap, err := loadCPCodesMap()
	if err != nil {
		return err
	}

	var codeReport []BytesDeliveredReport
	for key, codes := range cpCodeMap {
		report, err := r.getBytesDelivered(key, &codes, reportOptions)
		if err != nil {
			log.Printf("There was an error with the key: %s. Error: %v", key, err)
		}
		codeReport = append(codeReport, report...)

	}

	err = fileHandling.SaveToFile(codeReport, dataBasePath+"/result", "report.csv", "csv", delimiter[0])
	if err != nil {
		return err
	}

	return nil
}
