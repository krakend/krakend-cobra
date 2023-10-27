package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

type UnparsedConfigLoader interface {
	LoadUnparsed(cfgFile string) (interface{}, error)
}

func DefaultUnparserConfigLoader(cfgFile string) (interface{}, error) {
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("ERROR reading the configuration file %s: %s", cfgFile, err.Error())
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("ERROR reading JSON from the configuration file %s: %s", cfgFile, err.Error())
	}

	return raw, nil
}
