package job

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"fmt"

	"github.com/mleonard87/frosty/backup"
)

const (
	FROSTY_DIR_NAME                     = "frosty"
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

func getJobDirectoryPath(jobName string) string {
	userHome := getUserHomeDirectory()
	return filepath.Join(userHome, FROSTY_DIR_NAME, JOBS_DIR_NAME, jobName)
}

func getJobArtefactDirectoryPath(jobName string) string {
	userHome := getUserHomeDirectory()
	return filepath.Join(userHome, FROSTY_DIR_NAME, JOBS_DIR_NAME, jobName, JOB_ARTIFACTS_DIR_NAME)
}

func MakeJobDirectories(jobName string) (string, string, error) {
	jobDir := getJobDirectoryPath(jobName)
	artifactDir := getJobArtefactDirectoryPath(jobName)

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

func RemoveJobDirectory(jobName string) error {
	jobDir := getJobDirectoryPath(jobName)
	return os.RemoveAll(jobDir)

}

func GetArtifactArchiveFileName(jobName string) string {
	bs := *backupservice.CurrentBackupService()
	return fmt.Sprintf("%s.%s", bs.ArtifactFilename(jobName), ARTIFACT_ARCHIVE_FILENAME_EXTENSION)
}

func GetArtifactArchiveTargetName(jobName string) string {
	artifactDir := getJobArtefactDirectoryPath(jobName)
	return filepath.Join(artifactDir, GetArtifactArchiveFileName(jobName))
}
