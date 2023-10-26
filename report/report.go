package report

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/edgegrid"
	"github.com/jguerra6/akamai-reports/command"
	"github.com/jguerra6/akamai-reports/cpcode"
	"github.com/jguerra6/akamai-reports/fileHandling"
	"github.com/jguerra6/akamai-reports/request"
)

const (
	cpCodeMapFileName = "cpCode_list.json"
	dataBasePath      = "./data"
	dateFormat        = "2006-01-02T15:04:05Z"
	dateSuffix        = "T00:00:00Z"
	maxRequests       = 120
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

type Options struct {
	AccountSwitchKey string
	CpCodes          *[]cpcode.CPCode
	StartDate        string
	EndDate          string
	ReportConfig     map[string]string
}

func bytesDeliveredReports(edgerc *edgegrid.Config, requestUrl string, reportOptions map[string]string) ([]BytesDeliveredReport, error) {

	resp, err := request.AkamaiRequest(edgerc, requestUrl, reportOptions)
	if err != nil {
		return nil, err
	}
	reportResponse, err := parseResponse(resp, reportOptions)
	if err != nil {
		log.Printf("Error parsing the response. Err: %v", err)
		return nil, err
	}

	return reportResponse, nil

}

func reportRunner(workerID int, wg *sync.WaitGroup, taskChan chan map[string]string, edgerc *edgegrid.Config, requestUrl string, reports *[]BytesDeliveredReport, errors []string) {
	go func(workerID int) {
		defer wg.Done()
		for task := range taskChan {
			report, err := bytesDeliveredReports(edgerc, requestUrl, task)

			if err != nil {
				errors = append(errors, err.Error())
				continue
			}
			*reports = append(*reports, report...)
		}

	}(workerID)
}

// TODO:
// improve error logging
func (r *Report) getBytesDelivered(options *Options) ([]BytesDeliveredReport, error) {
	if *options.CpCodes == nil {
		return nil, nil
	}

	var (
		bytesDeliveredReport []BytesDeliveredReport
		errors               []string
		taskChan             = make(chan map[string]string)
		wg                   sync.WaitGroup
	)

	requestUrl := &url.URL{
		Path: request.ReplaceParams(options.ReportConfig, reportBaseUrl),
	}

	query := requestUrl.Query()
	query.Add("allObjectIds", "true")
	query.Add("accountSwitchKey", options.AccountSwitchKey)

	requestUrl.RawQuery = query.Encode()

	startDate, err := time.Parse(time.RFC3339, options.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := time.Parse(time.RFC3339, options.EndDate)
	if err != nil {
		return nil, err
	}

	for i := 0; i < maxRequests; i++ {
		wg.Add(1)
		reportRunner(i, &wg, taskChan, r.edgerc, requestUrl.RequestURI(), &bytesDeliveredReport, errors)
	}

	for startDate.Before(endDate) {
		reportOptions := map[string]string{}
		reportOptions["start"] = startDate.Format(dateFormat)
		startDate = startDate.AddDate(0, 0, 1)
		reportOptions["end"] = startDate.Format(dateFormat)

		taskChan <- reportOptions
	}

	close(taskChan)
	wg.Wait()

	//for _, s := range errors {
	//log.Printf("Error with key: %s, msg: %s", accountSwitchKey, s)
	//}
	startDate, err = time.Parse(time.RFC3339, options.StartDate)
	duration := endDate.Sub(startDate)

	// Calculate success rate for logging purposes
	days := duration.Hours() / 24.0
	daysInt := int(days)

	successCount := daysInt - len(errors)
	codesCount := daysInt
	errRate := float64(successCount) * 100.00 / float64(codesCount)
	log.Printf("%s completed: %d/%d = %f%%", options.AccountSwitchKey, successCount, codesCount, errRate)

	return bytesDeliveredReport, nil
}

func (r *Report) Run() error {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Printf("Akamai Reporting took %s", elapsed)
	}()

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

	if !strings.Contains(reportOptions["start"], dateSuffix) {
		reportOptions["start"] += dateSuffix
	}

	if !strings.Contains(reportOptions["end"], dateSuffix) {
		reportOptions["end"] += dateSuffix
	}

	cpCodeMap, err := loadCPCodesMap()
	if err != nil {
		return err
	}

	reportConfig := map[string]string{
		"reportName":    "bytes-by-cpcode",
		"reportVersion": "1",
	}

	var (
		codeReport []BytesDeliveredReport
		wg         sync.WaitGroup
	)

	resultChan := make(chan []BytesDeliveredReport)

	for key, codes := range cpCodeMap {
		wg.Add(1)

		go func(key string, codes []cpcode.CPCode) {
			defer wg.Done()
			options := &Options{
				AccountSwitchKey: key,
				CpCodes:          &codes,
				StartDate:        reportOptions["start"],
				EndDate:          reportOptions["end"],
				ReportConfig:     reportConfig,
			}
			report, err := r.getBytesDelivered(options)
			if err != nil {
				log.Printf("There was an error with the key: %s. Error: %v", key, err)
			}

			resultChan <- report
		}(key, codes)

	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for report := range resultChan {
		codeReport = append(codeReport, report...)
	}

	err = fileHandling.SaveToFile(codeReport, dataBasePath+"/result", "report.csv", "csv", delimiter[0])
	if err != nil {
		return err
	}

	return nil
}
