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
	ERROR_CODE_INVALID_BUCKET_NAME         string = "InvalidBucketName"
	ERROR_CODE_BUCKET_ALREADY_OWNED_BY_YOU string = "BucketAlreadyOwnedByYou"
	LIFECYCLE_ID                           string = "frosty-backup-retention-policy"
)

type AmazonS3BackupService struct {
	AccessKeyId        string
	SecretAccessKey    string
	Region             string
	AccountId          string
	RetentionDays      int64
	BucketName         string
	Endpoint           string
	UsePathStyleAccess bool
	S3Service          *s3.S3
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

	// Attempt to get the retentionDays config property. If this can't be found then default to 0.
	// 0 will not set a life cycle policy and any existing policy will remain.
	rd, ok := backupConfig.BackupConfig["retentionDays"]
	if ok {
		asbs.RetentionDays = int64(rd.(float64))
	} else {
		asbs.RetentionDays = 0
	}

	asbs.BucketName = backupConfig.BackupConfig["bucketName"].(string)

	// Endpoint is optional
	e, ok := backupConfig.BackupConfig["endpoint"]
	if ok {
		asbs.Endpoint = e.(string)
	} else {
		asbs.Endpoint = ""
	}

	// pathStyleAccess is optional
	psa, ok := backupConfig.BackupConfig["pathStyleAccess"]
	if ok {
		asbs.UsePathStyleAccess = psa.(bool)
	} else {
		asbs.UsePathStyleAccess = false
	}
}

// Initialise anything in the backup service that needs to be created prior to uploading files. In this instance we need
// to create a bucket to store the backups if one does not already exist. This always uses a bucket
// called "frosty.backups".
func (asbs *AmazonS3BackupService) Init() error {
	asbs.setEnvvars()

	ac := &aws.Config{}
	ac.S3ForcePathStyle = &asbs.UsePathStyleAccess
	if asbs.Endpoint != "" {
		ac.Endpoint = &asbs.Endpoint
	} else {
		ac = &aws.Config{}
	}

	asbs.S3Service = s3.New(session.New(), ac)

	err := asbs.createBucket(asbs.BucketName)
	if err != nil {
		log.Println("Error creating bucket")
		log.Println(err)
		return err
	}

	err = asbs.putBucketLifecycleConfiguration()
	if err != nil {
		log.Println("Error creating bucket lifecycle")
		log.Println(err)
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

func (asbs *AmazonS3BackupService) putBucketLifecycleConfiguration() error {
	// If the retention period is not 0 days then submit a new life cycle policy.
	if asbs.RetentionDays != 0 {
		params := &s3.PutBucketLifecycleConfigurationInput{
			Bucket: aws.String(asbs.BucketName),
			LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
				Rules: []*s3.LifecycleRule{
					{
						Prefix: aws.String(""),
						Status: aws.String("Enabled"),
						ID:     aws.String(LIFECYCLE_ID),
						Expiration: &s3.LifecycleExpiration{
							Days: &asbs.RetentionDays,
						},
					},
				},
			},
		}

		_, err := asbs.S3Service.PutBucketLifecycleConfiguration(params)
		if err != nil {
			log.Printf("Failed to create bucket lifecycle configuration, %s.\n", err)
			return err
		}
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
