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
	"github.com/mleonard87/frosty/reporting"
	"gopkg.in/robfig/cron.v2"
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

// Print usage information about the frosty backup tool.
func printHelp() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\n\tfrosty <path-to-frosty-config-file> [flags...]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n%s\n", frostyVersion)
}

// Prints version information about frosty.
func printVersion() {
	fmt.Fprintf(os.Stderr, "Frosty backup utility, version %s\n", frostyVersion)
}

// Validate a given config file.
func validate(configPath string) {
	_, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Frosty config file: %v - FAILED\n", configPath)
		os.Exit(1)
	}

	fmt.Printf("Frosty config file: %v - OK\n", configPath)
}

// The main function for beginning backups. This is the default way in which frosty will run. It loads a config file
// and then execute all the backups.
func backup(configPath string) {
	fc, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	bs := backupservice.NewBackupService(&fc.BackupConfig)
	sj := fc.ScheduledJobs()

	scheduleJobs(sj, bs, fc)

	select {}
}

// For the given map of cron schedule times against the list of jobs due to run at this time raise a gocron job
// to execute each of these jobs in go routines at the given time.
func scheduleJobs(js map[string][]config.JobConfig, bs backupservice.BackupService, fc config.FrostyConfig) {
	c := cron.New()

	for k, v := range js {
		// Assign v to jobs to use in the closure below.
		jobs := v

		// The function defined below acts as a closure using the assigned "jobs" variable above.
		// If we do not re-assign v to jobs as above and constantly used "v" in the function then
		// we would find that only the last set of jobs would ever be run as the value of "v" is
		// updated in each iteration of the loop. However, jobs is scoped within the body of the
		// loop and the closure below can take advantage of this.
		_, err := c.AddFunc(k, func() {
			// Get a timestamp as an ID for this run of jobs. This will be used in the directory name to ensure that
			// if jobs overlap we don't get any conflicts.
			t := time.Now()
			runId := t.Format("20060102150405")

			js := beginJobs(jobs, runId)
			initBackupService(bs, js)
			beginBackups(bs, js, runId)

			if &fc.ReportingConfig.Email != nil {
				reporting.SendEmailSummary(js, &fc.ReportingConfig.Email)
			}
		})

		if err != nil {
			log.Fatalf("Error scheduling jobs: %s", err.Error())
		}
	}

	c.Start()
}

// Starts running all jobs by executing the commands and letting each command create its artifacts. This function
// returns when all jobs have finished. Each job is run in a separate go routine.
func beginJobs(jobs []config.JobConfig, runId string) []job.JobStatus {
	ch := make(chan job.JobStatus)
	var wg sync.WaitGroup

	for _, j := range jobs {
		wg.Add(1)
		go beginJob(j, runId, ch, &wg)
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

// Run an individual job.
func beginJob(jobConfig config.JobConfig, runId string, ch chan job.JobStatus, wg *sync.WaitGroup) {
	log.Printf("Running Job: %s\n", jobConfig.Name)
	defer wg.Done()
	js := job.Start(jobConfig, runId)
	ch <- js
}

// Initialise the backup service (e.g. S3 or Glacier) if there was a problem doing this mark all jobs as failed and
// write the error message to each job.
func initBackupService(backupService backupservice.BackupService, jobStatuses []job.JobStatus) error {
	err := backupService.Init()

	if err != nil {
		for i := range jobStatuses {
			// If we couldn't init the backup service then just log the same error caused by that against
			// each job. This saves needing to create a generic section in the email reporting that covers
			// over-arching backup service errors.
			jobStatuses[i].Status = job.STATUS_FAILURE
			jobStatuses[i].TransferError = err.Error()
			continue
		}

		return err
	}

	return nil
}

// Begin the transfer of artifacts to the backup service.
func beginBackups(backupService backupservice.BackupService, jobStatuses []job.JobStatus, runId string) {
	for i, js := range jobStatuses {
		archivePath := job.GetArtifactArchiveTargetName(js.JobConfig.Name, runId)

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

		// Remove the directory created for this job.
		err = job.RemoveJobDirectory(js.JobConfig.Name, runId)
		if err != nil {
			em := fmt.Sprintf("Unable to remove working directory for %s job following successful transfer:\n%s\n", js.JobConfig.Name, err)
			js.Status = job.STATUS_FAILURE
			js.Error = em
			js.EndTime = time.Now()
			continue
		}
	}

	// Finally remove the run directory (this should be empty by this point).
	err := job.RemoveRunDirectory(runId)
	if err != nil {
		log.Printf("Error removing run directory \"%s\":\n%s\n", runId, err)
	}
}
