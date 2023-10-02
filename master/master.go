package master

import (
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/edgegrid"
	"github.com/jguerra6/akamai-reports/accountSwitch"
	"github.com/jguerra6/akamai-reports/command"
	"github.com/jguerra6/akamai-reports/cpcode"
	"github.com/jguerra6/akamai-reports/fileHandling"
)

const (
	dataBasePath      = "./data/master"
	switchKeyFileName = "switch_keys.json"
	cpCodeMapFileName = "cpCode_list.json"
)

type Master struct {
	*command.Command
	edgerc *edgegrid.Config
}

func New(edgerc *edgegrid.Config) *Master {
	return &Master{
		Command: command.NewCommand("init", nil),
		edgerc:  edgerc,
	}
}

func (m *Master) saveAccountSwitchKeys() ([]*accountSwitch.AccountSwitch, error) {
	switchKeys, err := accountSwitch.GetSwitchKeys(m.edgerc)
	if err != nil {
		return nil, err
	}

	err = fileHandling.SaveToFile(switchKeys, dataBasePath, switchKeyFileName, "json")
	if err != nil {
		return nil, err
	}

	return switchKeys, nil
}

func (m *Master) saveCPCodesMap(switchKeys []*accountSwitch.AccountSwitch) error {

	cpCodeMap, err := cpcode.GetAllCPCodes(m.edgerc, switchKeys)
	if err != nil {
		return err
	}

	err = fileHandling.SaveToFile(cpCodeMap, dataBasePath, cpCodeMapFileName, "json")
	if err != nil {
		return err
	}
	return nil
}

func (m *Master) Run() error {
	// Call this functions to create the CPCode maps
	switchKeys, err := m.saveAccountSwitchKeys()
	if err != nil {
		return err
	}
	// TODO: Add a flag check to read from the previous switchKey file
	//switchKeys, err := loadSwitchKeys()
	//if err != nil {
	//	return err
	//}

	err = m.saveCPCodesMap(switchKeys)
	if err != nil {
		return err
	}

	return nil
}
