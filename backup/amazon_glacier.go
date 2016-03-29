package backupservice

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/mleonard87/frosty/config"
)

type AmazonGlacierStorageService struct {
	AccessKeyId     string
	SecretAccessKey string
	Region          string
	AccountId       string
	VaultName       string
	GlacierService  *glacier.Glacier
}

func (agss *AmazonGlacierStorageService) SetConfig(backupConfig *config.BackupConfig) {
	agss.AccessKeyId = backupConfig.AmazonGlacierBackupConfig.AccessKeyId
	agss.SecretAccessKey = backupConfig.AmazonGlacierBackupConfig.SecretAccessKey
	agss.Region = backupConfig.AmazonGlacierBackupConfig.Region
	agss.AccountId = backupConfig.AmazonGlacierBackupConfig.AccountId
	agss.VaultName = GetBackupName()
}

func (agss *AmazonGlacierStorageService) Init() {
	agss.setEnvvars()
	agss.GlacierService = glacier.New(session.New(), &aws.Config{})

	//params := &glacier.ListVaultsInput{
	//	AccountId: aws.String(agss.AccountId),
	//}

	//resp, err := agss.GlacierService.ListVaults(params)

	//if err != nil {
	//	// Print the error, cast err to awserr.Error to get the Code and
	//	// Message from an error.
	//	fmt.Println(err)
	//	return
	//}

	agss.createVault(agss.VaultName)

	// Pretty-print the response data.
	//fmt.Println(resp)
}

func (agss *AmazonGlacierStorageService) StoreFile(pathToFile string) {
	fileContents, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		log.Fatal(err)
	}

	params := &glacier.UploadArchiveInput{
		AccountId: aws.String(agss.AccountId),
		VaultName: aws.String(agss.VaultName),
		Body:      bytes.NewReader(fileContents),
	}
	resp, err := agss.GlacierService.UploadArchive(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}

func (agss *AmazonGlacierStorageService) createVault(vaultName string) {
	agss.setEnvvars()
	params := &glacier.CreateVaultInput{
		AccountId: aws.String(agss.AccountId),
		VaultName: aws.String(agss.VaultName),
	}
	resp, err := agss.GlacierService.CreateVault(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err)
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}

func (agss *AmazonGlacierStorageService) setEnvvars() {
	os.Setenv("AWS_ACCESS_KEY_ID", agss.AccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", agss.SecretAccessKey)
	os.Setenv("AWS_REGION", agss.Region)
}
