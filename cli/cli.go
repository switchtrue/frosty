package cli

import (
	"fmt"
	"os"

	"sync"

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
	config.LoadConfig(configPath)
	fmt.Println("Frosty config file: %v - OK", configPath)
}

func backup(configPath string) {
	frostyConfig := config.LoadConfig(configPath)

	beginJobs(frostyConfig.Jobs)
}

func beginJobs(jobs []config.JobConfig) {

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

	for js := range ch {
		fmt.Println("Completed " + js.JobConfig.Name)
	}

}

func beginJob(jobConfig config.JobConfig, ch chan job.JobStatus, wg *sync.WaitGroup) {
	defer wg.Done()
	js := job.Start(jobConfig)
	ch <- js
}
