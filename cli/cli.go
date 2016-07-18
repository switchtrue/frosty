package cli

import (
	"fmt"
	"log"
	"os"

	"sync"

	"time"

	flag "github.com/ogier/pflag"

	"github.com/mleonard87/frosty/backup"
	"github.com/mleonard87/frosty/config"
	"github.com/mleonard87/frosty/job"
)

var frostyVersion string

const (
	COMMAND_BACKUP   = "backup"
	COMMAND_HELP     = "help"
	COMMAND_VALIDATE = "validate"
	COMMAND_VERSION  = "version"
)

func Execute() {
	flag.Usage = printHelp

	doValidate := flag.Bool("validate", false, "Validates that the specified config file is valid.")
	doVersion := flag.Bool("version", false, "Prints the version information about the Frosty backup utility.")

	flag.Parse()

	switch {
	case *doValidate:
		validate(os.Args[1])
	case *doVersion:
		printVersion()
	default:
		backup(os.Args[1])
	}
}

func printHelp() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\n\tfrosty <path-to-frosty-config-file> [flags...]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n%s\n", frostyVersion)
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

	backupService := backupservice.NewBackupService(&frostyConfig.BackupConfig)
	initBackupService(backupService, jobStatuses)
	//beginBackups(backupService, jobStatuses)

	//if &frostyConfig.ReportingConfig.Email != nil {
	//	reporting.SendEmailSummary(jobStatuses, &frostyConfig.ReportingConfig.Email)
	//}
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

func initBackupService(backupService backupservice.BackupService, jobStatuses []job.JobStatus) error {
	err := backupService.Init()

	if err != nil {
		for i := range jobStatuses {
			// If we couldn't init the backup service then just log the same error caused by that against each job.
			// this saves needing to create a generic section in the email reporting that covers over-arching
			// backup service errors.
			jobStatuses[i].Status = job.STATUS_FAILURE
			jobStatuses[i].TransferError = err.Error()
			continue
		}

		return err
	}

	return nil
}

func beginBackups(backupService backupservice.BackupService, jobStatuses []job.JobStatus) {
	for i, js := range jobStatuses {
		archivePath := job.GetArtifactArchiveTargetName(js.JobConfig.Name)

		// Only run the backup if the archive exists.
		_, err := os.Stat(archivePath)
		if err != nil {
			if !os.IsNotExist(err) {
				jobStatuses[i].Status = job.STATUS_FAILURE
				jobStatuses[i].TransferError = err.Error()
				continue
			} else {
				continue
			}
		}

		jobStatuses[i].TransferStartTime = time.Now()
		err = backupService.StoreFile(archivePath)
		if err != nil {
			jobStatuses[i].Status = job.STATUS_FAILURE
			jobStatuses[i].TransferError = err.Error()
			continue
		}
		jobStatuses[i].TransferEndTime = time.Now()
	}
}
