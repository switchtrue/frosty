package backupservice

import (
	"os"

	"errors"

	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mleonard87/frosty/config"
)

type AmazonS3BackupService struct {
	AccessKeyId     string
	SecretAccessKey string
	Region          string
	AccountId       string
	BucketName      string
	S3Service       *s3.S3
}

func (asbs *AmazonS3BackupService) SetConfig(backupConfig *config.BackupConfig) {
	asbs.AccessKeyId = backupConfig.BackupConfig["accessKeyId"].(string)
	asbs.SecretAccessKey = backupConfig.BackupConfig["secretAccessKey"].(string)
	asbs.Region = backupConfig.BackupConfig["region"].(string)
	asbs.AccountId = backupConfig.BackupConfig["accountId"].(string)
	asbs.BucketName = GetBackupName()
}

func (asbs *AmazonS3BackupService) Init() error {
	asbs.setEnvvars()
	asbs.S3Service = s3.New(session.New(), &aws.Config{})

	err := asbs.createBucket(asbs.BucketName)
	if err != nil {
		return err
	}

	return nil
}

func (asbs *AmazonS3BackupService) StoreFile(pathToFile string) error {
	return errors.New("not implemented - cannot store files")
}

func (asbs *AmazonS3BackupService) createBucket(bucketName string) error {
	asbs.setEnvvars()
	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	resp, err := asbs.S3Service.CreateBucket(params)
	if err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Println(resp)

	return nil
}

func (asbs *AmazonS3BackupService) setEnvvars() {
	os.Setenv("AWS_ACCESS_KEY_ID", asbs.AccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", asbs.SecretAccessKey)
	os.Setenv("AWS_REGION", asbs.Region)
}
