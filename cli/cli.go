package cli

import (
	"fmt"
	"log"
	"os"

	"sync"

	"github.com/mleonard87/frosty/config"
	"github.com/mleonard87/frosty/job"
	"github.com/mleonard87/frosty/reporting"
)

var frostyVersion string

const (
	COMMAND_BACKUP   = "backup"
	COMMAND_HELP     = "help"
	COMMAND_VALIDATE = "validate"
	COMMAND_VERSION  = "version"
)

func Execute() {
	switch os.Args[1] {
	case COMMAND_BACKUP:
		backup(os.Args[2])
	case COMMAND_HELP:
		printHelp()
	case COMMAND_VALIDATE:
		validate(os.Args[2])
	case COMMAND_VERSION:
		printVersion()
	}
}

func printHelp() {
	fmt.Println("usage: frosty <command> [<path-to-frosty-config-file>]")

	printVersion()

	fmt.Println("")
	fmt.Println("Available commands:")
	fmt.Printf("  %s <path-to-frosty-config-file> - executes a backup for the specified config file\n", COMMAND_BACKUP)
	fmt.Printf("  %s - prints this help information\n", COMMAND_HELP)
	fmt.Printf("  %s <path-to-frosty-config-file> - validates that the specified config file\n", COMMAND_VALIDATE)
	fmt.Printf("  %s - prints version information about the Frosty backup utility\n", COMMAND_VERSION)
}

func printVersion() {
	fmt.Printf("Frosty backup utility, version %s\n", frostyVersion)
}

func validate(configPath string) {
	_, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Frosty config file: %v - FAILED\n", configPath)
		os.Exit(1)
	}

	fmt.Printf("Frosty config file: %v - OK\n", configPath)
}

func backup(configPath string) {
	frostyConfig, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	jobStatuses := beginJobs(frostyConfig.Jobs)

	if &frostyConfig.ReportingConfig.Email != nil {
		reporting.SendEmailSummary(jobStatuses, &frostyConfig.ReportingConfig.Email)
	}
}

func beginJobs(jobs []config.JobConfig) []job.JobStatus {

	ch := make(chan job.JobStatus)
	var wg sync.WaitGroup

	for _, j := range jobs {
		wg.Add(1)
		go beginJob(j, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var jobStatuses []job.JobStatus

	for js := range ch {
		jobStatuses = append(jobStatuses, js)
	}

	return jobStatuses
}

func beginJob(jobConfig config.JobConfig, ch chan job.JobStatus, wg *sync.WaitGroup) {
	defer wg.Done()
	js := job.Start(jobConfig)
	ch <- js
}
