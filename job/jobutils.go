package job

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"fmt"

	"github.com/mleonard87/frosty/backup"
	"github.com/mleonard87/frosty/config"
)

const (
	FROSTY_DIR_NAME                     = ".frosty"
	JOBS_DIR_NAME                       = "jobs"
	JOB_ARTIFACTS_DIR_NAME              = "artifacts"
	ARTIFACT_ARCHIVE_FILENAME_EXTENSION = "zip"
)

func getUserHomeDirectory() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to obtain current user: %s", err)
		os.Exit(1)
	}

	return usr.HomeDir
}

func getRunDirectoryPath(runId string) string {
	fc := config.GetFrostConfig()
	if fc.WorkDir == "" {
		userHome := getUserHomeDirectory()
		return filepath.Join(userHome, FROSTY_DIR_NAME, JOBS_DIR_NAME, runId)
	} else {
		return filepath.Join(fc.WorkDir, JOBS_DIR_NAME, runId)
	}
}

func getJobDirectoryPath(jobName string, runId string) string {
	return filepath.Join(getRunDirectoryPath(runId), jobName)
}

func getJobArtifactDirectoryPath(jobName string, runId string) string {
	return filepath.Join(getJobDirectoryPath(jobName, runId), JOB_ARTIFACTS_DIR_NAME)
}

func MakeJobDirectories(jobName string, runId string) (string, string, error) {
	jobDir := getJobDirectoryPath(jobName, runId)
	artifactDir := getJobArtifactDirectoryPath(jobName, runId)

	err := os.MkdirAll(jobDir, 0755)
	if err != nil {
		return "", "", err
	}

	err2 := os.MkdirAll(artifactDir, 0755)
	if err2 != nil {
		return "", "", err
	}

	return jobDir, artifactDir, nil
}

func RemoveJobDirectory(jobName string, runId string) error {
	jobDir := getJobDirectoryPath(jobName, runId)
	return os.RemoveAll(jobDir)
}

func RemoveRunDirectory(runId string) error {
	runDir := getRunDirectoryPath(runId)
	return os.RemoveAll(runDir)
}

func GetArtifactArchiveFileName(jobName string) string {
	bs := *backupservice.CurrentBackupService()
	return fmt.Sprintf("%s.%s", bs.ArtifactFilename(jobName), ARTIFACT_ARCHIVE_FILENAME_EXTENSION)
}

func GetArtifactArchiveTargetName(jobName string, runId string) string {
	artifactDir := getJobArtifactDirectoryPath(jobName, runId)
	return filepath.Join(artifactDir, GetArtifactArchiveFileName(jobName))
}
