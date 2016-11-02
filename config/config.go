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
	BACKUP_SERVICE_AMAZON_GLACIER = "glacier"
	BACKUP_SERVICE_AMAZON_S3      = "s3"
)

var frostyConfig FrostyConfig

type FrostyConfig struct {
	ReportingConfig ReportingConfig        `json:"reporting"`
	RawBackupConfig map[string]interface{} `json:"backup"`
	BackupConfig    BackupConfig
	Jobs            []JobConfig `json:"jobs"`
}

type ReportingConfig struct {
	Email EmailReportingConfig `json:"email"`
}

type EmailReportingConfig struct {
	SMTP struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"smtp"`
	Sender     string   `json:"sender"`
	Recipients []string `json:"recipients"`
}

type JobConfig struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	Schedule string `json:"schedule"`
}

type BackupConfig struct {
	BackupService string
	BackupConfig  map[string]interface{}
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
	// TODO: Validate that the schedules passed in are valid cron syntax.

	return validationPassed
}

func (fc *FrostyConfig) ScheduledJobs() map[string][]JobConfig {
	sj := make(map[string][]JobConfig)

	for _, v := range fc.Jobs {

		val, ok := sj[v.Schedule]
		if !ok {
			val = []JobConfig{}
		}

		val = append(val, v)

		sj[v.Schedule] = val
	}

	return sj
}

func LoadConfig(configPath string) (FrostyConfig, error) {
	f, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Cannot find frosty config file: %s\n", err)
		os.Exit(1)
	}

	var fc FrostyConfig
	err = json.Unmarshal(f, &fc)
	if err != nil {
		log.Fatalf("Cannot parse frosty config file: %v: %v Are you sure the JSON is valid?\n", configPath, err)
		os.Exit(1)
	}

	var backupConfig BackupConfig
	if config, ok := fc.RawBackupConfig[BACKUP_SERVICE_AMAZON_GLACIER]; ok {
		backupConfig.BackupService = BACKUP_SERVICE_AMAZON_GLACIER
		backupConfig.BackupConfig = config.(map[string]interface{})
	}

	if config, ok := fc.RawBackupConfig[BACKUP_SERVICE_AMAZON_S3]; ok {
		backupConfig.BackupService = BACKUP_SERVICE_AMAZON_S3
		backupConfig.BackupConfig = config.(map[string]interface{})
	}

	fc.BackupConfig = backupConfig

	if !fc.validate() {
		return fc, errors.New("Failed to validate config file: " + configPath)
	}

	frostyConfig = fc

	return fc, nil
}

func GetFrostConfig() FrostyConfig {
	fc := frostyConfig
	return fc
}
