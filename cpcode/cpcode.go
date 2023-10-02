package cpcode

import (
	"encoding/json"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/edgegrid"
	"github.com/jguerra6/akamai-reports/accountSwitch"
	"github.com/jguerra6/akamai-reports/request"
)

const (
	baseUrl = "/cprg/v1/cpcodes"
)

type Root struct {
	CPCodes []CPCode `json:"cpcodes"`
}

type CPCode struct {
	AccountId string `json:"accountId"`
	CPCodeId  int    `json:"cpcodeId"`
	Name      string `json:"cpcodeName"`
}

func GetCpCodes(edgerc *edgegrid.Config, switchKey string) ([]CPCode, error) {
	requestUrl := baseUrl

	requestOptions := map[string]string{
		"accountSwitchKey": switchKey,
	}

	resp, err := request.AkamaiRequest(edgerc, requestUrl, requestOptions)
	if err != nil {
		return nil, err
	}

	var cpCodes Root

	err = json.Unmarshal(resp, &cpCodes)
	if err != nil {
		return nil, err
	}

	var cpCodeMap []CPCode

	for _, cpCode := range cpCodes.CPCodes {
		cpCodeMap = append(cpCodeMap, cpCode)
	}

	return cpCodeMap, nil
}

func GetAllCPCodes(edgerc *edgegrid.Config, switchKeys []*accountSwitch.AccountSwitch) (map[string][]CPCode, error) {

	cpCodeMap := map[string][]CPCode{}
	var (
		err error
	)

	for _, switchKey := range switchKeys {
		// This account has too many CP Codes causing a failure, currently it's being skipped but should be somehow handled
		if switchKey.AccountName == "Brightcove Inc._Value Added Reseller" {
			continue
		}
		key := switchKey.AccountSwitchKey
		cpCodeMap[key], err = GetCpCodes(edgerc, key)
		if err != nil {
			return nil, err
		}

	}

	return cpCodeMap, nil

}
