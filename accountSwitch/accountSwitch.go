package accountSwitch

import (
	"encoding/json"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/edgegrid"
	"github.com/jguerra6/akamai-reports/request"
)

const (
	baseUrl = "/identity-management/v3"
)

type AccountSwitch struct {
	AccountSwitchKey string `json:"accountSwitchKey"`
	AccountName      string `json:"accountName"`
}

func GetSwitchKeys(edgerc *edgegrid.Config) ([]*AccountSwitch, error) {
	requestUrl := baseUrl + "/api-clients/self/account-switch-keys"
	resp, err := request.AkamaiRequest(edgerc, requestUrl, nil)
	if err != nil {
		return nil, err
	}

	var switchKeys []*AccountSwitch

	err = json.Unmarshal(resp, &switchKeys)
	if err != nil {
		return nil, err
	}

	return switchKeys, nil
}
