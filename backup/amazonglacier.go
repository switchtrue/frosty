package backupservice

import (
	"bytes"
	"fmt"
	"os"

	"io/ioutil"

	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/mleonard87/frosty/config"
)

const ()

type AmazonGlacierBackupService struct {
	AccessKeyId     string
	SecretAccessKey string
	Region          string
	AccountId       string
	VaultName       string
	GlacierService  *glacier.Glacier
}

// Return the backup service type this must match the string as used as the JSON property in the frosty backup config.
func (agss *AmazonGlacierBackupService) Name() string {
	return config.BACKUP_SERVICE_AMAZON_GLACIER
}

// Initialise any variable needed for backups.
func (agss *AmazonGlacierBackupService) SetConfig(backupConfig *config.BackupConfig) {
	agss.AccessKeyId = backupConfig.BackupConfig["accessKeyId"].(string)
	agss.SecretAccessKey = backupConfig.BackupConfig["secretAccessKey"].(string)
	agss.Region = backupConfig.BackupConfig["region"].(string)
	agss.AccountId = backupConfig.BackupConfig["accountId"].(string)
	agss.VaultName = agss.getVaultName()
}

// Initialise anything in the backup service that needs to be created prior to uploading files. In this instance we need
// to create a vault for the backup to hold any archives.
func (agss *AmazonGlacierBackupService) Init() error {
	agss.setEnvvars()
	agss.GlacierService = glacier.New(session.New(), &aws.Config{})

	err := agss.createVault(agss.VaultName)
	if err != nil {
		return err
	}

	return nil
}

// Store the file in pathToFile in Amazon Glacier.
func (agss *AmazonGlacierBackupService) StoreFile(pathToFile string) error {
	fileContents, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		return err
	}

	params := &glacier.UploadArchiveInput{
		AccountId: aws.String(agss.AccountId),
		VaultName: aws.String(agss.VaultName),
		Body:      bytes.NewReader(fileContents),
	}

	_, err = agss.GlacierService.UploadArchive(params)
	if err != nil {
		return err
	}

	return nil
}

// Get the name to be used for the .zip archive without the .zip extension.
func (agss *AmazonGlacierBackupService) ArtifactFilename(jobName string) string {
	return jobName
}

// Get a friendly name for the email template of where this backup was stored. In this case, the name of the Glacier
// vault.
func (agss *AmazonGlacierBackupService) BackupLocation() string {
	return fmt.Sprintf("Glacier Vault: %s", agss.getVaultName())
}

// Create the Glacier Vault.
func (agss *AmazonGlacierBackupService) createVault(vaultName string) error {
	agss.setEnvvars()
	params := &glacier.CreateVaultInput{
		AccountId: aws.String(agss.AccountId),
		VaultName: aws.String(agss.VaultName),
	}

	_, err := agss.GlacierService.CreateVault(params)
	if err != nil {
		return err
	}

	return nil
}

// Set the required AWS environment variables.
func (agss *AmazonGlacierBackupService) setEnvvars() {
	os.Setenv(ENVVAR_AWS_ACCESS_KEY_ID, agss.AccessKeyId)
	os.Setenv(ENVVAR_AWS_SECRET_ACCESS_KEY, agss.SecretAccessKey)
	os.Setenv(ENVVAR_AWS_REGION, agss.Region)
}

// Get the name to use for for the Glacier Vault.
func (agss *AmazonGlacierBackupService) getVaultName() string {
	if agss.VaultName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal("Could not determine hostname.", err)
		}
		agss.VaultName = BACKUP_NAME_PREFIX + time.Now().Format("20060102") + "_" + hostname
	}

	return agss.VaultName
}
