package job

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

const (
	FROSTY_DIR_NAME        = "frosty"
	JOBS_DIR_NAME          = "jobs"
	JOB_ARTIFACTS_DIR_NAME = "artifacts"
)

func getUserHomeDirectory() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to obtain current user: %s", err)
		os.Exit(1)
	}

	return usr.HomeDir
}

func getJobDirectoryPath(jobName string) string {
	userHome := getUserHomeDirectory()
	return filepath.Join(userHome, FROSTY_DIR_NAME, JOBS_DIR_NAME, jobName)
}

func getJobArtefactDirectoryPath(jobName string) string {
	userHome := getUserHomeDirectory()
	return filepath.Join(userHome, FROSTY_DIR_NAME, JOBS_DIR_NAME, jobName, JOB_ARTIFACTS_DIR_NAME)
}

func MakeJobDirectories(jobName string) (string, string) {
	jobDir := getJobDirectoryPath(jobName)
	artefactDir := getJobArtefactDirectoryPath(jobName)

	err := os.MkdirAll(jobDir, 0755)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err2 := os.MkdirAll(artefactDir, 0755)
	if err2 != nil {
		log.Fatal(err2)
		os.Exit(1)
	}

	return jobDir, artefactDir
}
