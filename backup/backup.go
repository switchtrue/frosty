package backupservice

import (
	"log"
	"os"
	"time"

	"github.com/mleonard87/frosty/config"
)

const (
	BACKUP_NAME_PREFIX = "frosty_"
)

type BackupService interface {
	SetConfig(backupConfig *config.BackupConfig)
	Init() error
	StoreFile(pathToFile string) error
}

func NewBackupService(backupConfig *config.BackupConfig) BackupService {
	var backupService BackupService
	switch backupConfig.BackupService {
	case config.BACKUP_SERVICE_AMAZON_GLACIER:
		backupService = &AmazonGlacierBackupService{}
	default:
		log.Fatal("Only Amazon Glacier is supported as a backup service.")
		return nil
	}

	backupService.SetConfig(backupConfig)

	return backupService
}

var vaultName string

func GetBackupName() string {
	if vaultName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal("Could not determine hostname.", err)
		}
		vaultName = BACKUP_NAME_PREFIX + time.Now().Format("20060102") + "_" + hostname
	}

	return vaultName
}
