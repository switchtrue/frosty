package backupservice

import (
	"bytes"
	"os"

	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/mleonard87/frosty/config"
)

type AmazonGlacierBackupService struct {
	AccessKeyId     string
	SecretAccessKey string
	Region          string
	AccountId       string
	VaultName       string
	GlacierService  *glacier.Glacier
}

func (agss *AmazonGlacierBackupService) SetConfig(backupConfig *config.BackupConfig) {
	agss.AccessKeyId = backupConfig.BackupConfig["accessKeyId"].(string)
	agss.SecretAccessKey = backupConfig.BackupConfig["secretAccessKey"].(string)
	agss.Region = backupConfig.BackupConfig["region"].(string)
	agss.AccountId = backupConfig.BackupConfig["accountId"].(string)
	agss.VaultName = GetBackupName()
}

func (agss *AmazonGlacierBackupService) Init() error {
	agss.setEnvvars()
	agss.GlacierService = glacier.New(session.New(), &aws.Config{})

	err := agss.createVault(agss.VaultName)
	if err != nil {
		return err
	}

	return nil
}

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

func (agss *AmazonGlacierBackupService) setEnvvars() {
	os.Setenv("AWS_ACCESS_KEY_ID", agss.AccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", agss.SecretAccessKey)
	os.Setenv("AWS_REGION", agss.Region)
}
