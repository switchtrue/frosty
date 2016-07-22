package backupservice

import (
	"log"
	"os"

	"fmt"

	"path/filepath"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mleonard87/frosty/config"
)

const (
	BUCKET_NAME                            string = "frosty.backups"
	ERROR_CODE_INVALID_BUCKET_NAME         string = "InvalidBucketName"
	ERROR_CODE_BUCKET_ALREADY_OWNED_BY_YOU string = "BucketAlreadyOwnedByYou"
)

type AmazonS3BackupService struct {
	AccessKeyId     string
	SecretAccessKey string
	Region          string
	AccountId       string
	BucketName      string
	S3Service       *s3.S3
}

// Return the backup service type this must match the string as used as the JSON property in the frosty backup config.
func (agss *AmazonS3BackupService) Name() string {
	return config.BACKUP_SERVICE_AMAZON_S3
}

// Initialise any variable needed for backups.
func (asbs *AmazonS3BackupService) SetConfig(backupConfig *config.BackupConfig) {
	asbs.AccessKeyId = backupConfig.BackupConfig["accessKeyId"].(string)
	asbs.SecretAccessKey = backupConfig.BackupConfig["secretAccessKey"].(string)
	asbs.Region = backupConfig.BackupConfig["region"].(string)
	asbs.AccountId = backupConfig.BackupConfig["accountId"].(string)
	asbs.BucketName = BUCKET_NAME
}

// Initialise anything in the backup service that needs to be created prior to uploading files. In this instance we need
// to create a bucket to store the backups if one does not already exist. This always uses a bucket
// called "frosty.backups".
func (asbs *AmazonS3BackupService) Init() error {
	asbs.setEnvvars()
	asbs.S3Service = s3.New(session.New(), &aws.Config{})

	err := asbs.createBucket(asbs.BucketName)
	if err != nil {
		fmt.Println("error creating bucket")
		fmt.Println(err)
		return err
	}

	return nil
}

// Store the file in pathToFile in the bucket in S3.
func (asbs *AmazonS3BackupService) StoreFile(pathToFile string) error {
	_, fileName := filepath.Split(pathToFile)

	key := getObjectKey(fileName)

	f, err := os.Open(pathToFile)
	if err != nil {
		log.Printf("Failed to open file to store: %s", pathToFile)
		log.Println(err)
		return err
	}

	params := &s3.PutObjectInput{
		Body:   f,
		Bucket: &asbs.BucketName,
		Key:    &key,
	}

	_, err = asbs.S3Service.PutObject(params)
	if err != nil {
		log.Printf("Failed to put object %s into bucket %s with a key of %s\n", pathToFile, asbs.BucketName, fileName)
		log.Println(err)
		return err
	}

	return nil
}

// Get the name to be used for the .zip archive without the .zip extension.
func (asbs *AmazonS3BackupService) ArtifactFilename(jobName string) string {
	return jobName
}

// Get a friendly name for the email template of where this backup was stored. In this case, the name of the S3 bucket.
func (asbs *AmazonS3BackupService) BackupLocation() string {
	return fmt.Sprintf("S3 Bucket: %s", asbs.BucketName)
}

// Create the S3 bucket.
func (asbs *AmazonS3BackupService) createBucket(bucketName string) error {
	asbs.setEnvvars()
	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	_, err := asbs.S3Service.CreateBucket(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == ERROR_CODE_INVALID_BUCKET_NAME {
				log.Printf("The specified bucket is not valid. Bucket name: %s\n", bucketName)
				log.Println(err)
				return err
			}
			// If the BucketAlreadyOwnedByYou error is raised then this bucket already exists.
			if aerr.Code() == ERROR_CODE_BUCKET_ALREADY_OWNED_BY_YOU {
				return nil
			}

		}
		return err
	}

	if err = asbs.S3Service.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: &bucketName}); err != nil {
		log.Printf("Failed to wait for bucket to exist %s, %s\n", bucketName, err)
		return err
	}

	return nil
}

// Set the required AWS environment variables.
func (asbs *AmazonS3BackupService) setEnvvars() {
	os.Setenv(ENVVAR_AWS_ACCESS_KEY_ID, asbs.AccessKeyId)
	os.Setenv(ENVVAR_AWS_SECRET_ACCESS_KEY, asbs.SecretAccessKey)
	os.Setenv(ENVVAR_AWS_REGION, asbs.Region)
}

// Get the name to use for the file being stored in S3.
func getObjectKey(fileName string) string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Could not determine hostname.", err)
		return ""
	}

	return fmt.Sprintf("%s/%s/%s_%s", hostname, time.Now().Format("20060102"), time.Now().Format("15:04:05"), fileName)
}
