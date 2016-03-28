package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type FrostyConfig struct {
	ReportingConfig ReportingConfig `json:"reporting"`
	Jobs            []JobConfig     `json:"jobs"`
}

type ReportingConfig struct {
	Email EmailReportingConfig `json:"email"`
}

type EmailReportingConfig struct {
	SMTP struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"smtp"`
	Sender     string   `json:"sender"`
	Recipients []string `json:"recipients"`
}

type JobConfig struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

func LoadConfig(configPath string) FrostyConfig {
	f, ferr := ioutil.ReadFile(configPath)
	if ferr != nil {
		log.Fatal("Cannot find frosty config file: %v\n", ferr)
		os.Exit(1)
	}

	var frostyConfig FrostyConfig
	jerr := json.Unmarshal(f, &frostyConfig)
	if jerr != nil {
		log.Fatal("Cannot parse frosty config file: %v: %v\n", configPath, ferr)
		os.Exit(1)
	}

	return frostyConfig
}
