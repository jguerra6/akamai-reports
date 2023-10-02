package request

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/edgegrid"
)

func ReplaceParams(params map[string]string, baseString string) string {
	resultString := baseString
	for key, value := range params {
		resultString = strings.ReplaceAll(resultString, "{{"+key+"}}", value)
	}
	return resultString
}

func AkamaiRequest(edgerc *edgegrid.Config, url string, params map[string]string) ([]byte, error) {

	client := http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	for key, param := range params {
		query.Add(key, param)
	}

	req.URL.RawQuery = query.Encode()
	edgerc.SignRequest(req)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	return body, nil

}
