package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	BACKUP_SERVICE_AMAZON_GLACIER = iota
)

type FrostyConfig struct {
	ReportingConfig ReportingConfig `json:"reporting"`
	BackupConfig    BackupConfig    `json:"backup"`
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

type BackupConfig struct {
	BackupService             int
	AmazonGlacierBackupConfig AmazonGlacierBackupConfig `json:"glacier"`
}

type AmazonGlacierBackupConfig struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
	AccountId       string `json:"accountId"`
}

func (fc *FrostyConfig) validateJobNames() bool {
	ok := true
	for _, j := range fc.Jobs {
		jobNameCount := 0
		for _, oj := range fc.Jobs {
			if j.Name == oj.Name {
				jobNameCount++
			}
		}
		if jobNameCount > 1 {
			ok = false
			log.Printf("Job names must be unique - duplicate found for %q.\n", j.Name)
		}
	}
	return ok
}

func (fc *FrostyConfig) validateJobs() bool {
	ok := true
	for i, j := range fc.Jobs {
		if strings.TrimSpace(j.Name) == "" {
			log.Printf("All jobs must have names and it must not be empty - job in position %d has no name.", i)
			ok = false
		}
		if strings.TrimSpace(j.Command) == "" {
			log.Printf("All jobs must have a command and it must not be empty - %q has no command.", j.Name)
			ok = false
		}
	}
	return ok
}

func (fc *FrostyConfig) validate() bool {
	validationPassed := true
	validationPassed = fc.validateJobNames()
	validationPassed = fc.validateJobs()

	// TODO: Validate that if the email section is supplied then all the details are provided.
	// TODO: Validate that the email addresses in the email section are actually email addresses.

	return validationPassed
}

func LoadConfig(configPath string) (FrostyConfig, error) {
	f, ferr := ioutil.ReadFile(configPath)
	if ferr != nil {
		log.Fatal("Cannot find frosty config file: %v\n", ferr)
		os.Exit(1)
	}

	var frostyConfig FrostyConfig
	jerr := json.Unmarshal(f, &frostyConfig)
	if jerr != nil {
		log.Fatal("Cannot parse frosty config file: %v: %v Are you sure the JSON is valid?\n", configPath, ferr)
		os.Exit(1)
	}
	// Hard coded for now as Amazon Glacier is the only service I have anticipated supporting
	frostyConfig.BackupConfig.BackupService = BACKUP_SERVICE_AMAZON_GLACIER

	if !frostyConfig.validate() {
		return frostyConfig, errors.New("Failed to validate config file: " + configPath)
	}

	return frostyConfig, nil
}
