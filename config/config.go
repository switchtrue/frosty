package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type FrostyConfig struct {
	WorkingDirectory string      `json:"workingDirectory"`
	Jobs             []JobConfig `json:"jobs"`
}

type JobConfig struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

func LoadConfig(configPath string) FrostyConfig {
	f, ferr := ioutil.ReadFile(configPath)
	if ferr != nil {
		fmt.Printf("Cannot find frosty config file: %v\n", ferr)
		os.Exit(1)
	}

	var frostyConfig FrostyConfig
	jerr := json.Unmarshal(f, &frostyConfig)
	if jerr != nil {
		fmt.Printf("Cannot parse frosty config file: %v: %v\n", configPath, ferr)
		os.Exit(1)
	}

	return frostyConfig
}
