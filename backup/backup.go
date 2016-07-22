package backupservice

import (
	"log"

	"github.com/mleonard87/frosty/config"
)

const (
	BACKUP_NAME_PREFIX           = "frosty_"
	ENVVAR_AWS_ACCESS_KEY_ID     = "AWS_ACCESS_KEY_ID"
	ENVVAR_AWS_SECRET_ACCESS_KEY = "AWS_SECRET_ACCESS_KEY"
	ENVVAR_AWS_REGION            = "AWS_REGION"
)

type BackupService interface {
	Name() string
	SetConfig(backupConfig *config.BackupConfig)
	Init() error
	StoreFile(pathToFile string) error
	ArtifactFilename(jobName string) string
	BackupLocation() string
}

var currentBackupService BackupService

func NewBackupService(backupConfig *config.BackupConfig) BackupService {
	var bs BackupService

	switch backupConfig.BackupService {
	case BACKUP_SERVICE_AMAZON_GLACIER:
		bs = &AmazonGlacierBackupService{}
	case BACKUP_SERVICE_AMAZON_S3:
		bs = &AmazonS3BackupService{}
	default:
		log.Fatal("Only Amazon Glacier and Amazon S3 are supported as a backup services.")
		return nil
	}

	bs.SetConfig(backupConfig)

	currentBackupService = bs

	return bs
}

func CurrentBackupService() *BackupService {
	if currentBackupService == nil {
		log.Fatal("You must create a backup service before calling CurrentbackupService().")
	}
	return &currentBackupService
}
