# Frosty Backup Utility

A lightweight command line backup utility that stores back ups as archives in Amazon Glacier or S3.

Frosty will run one or more user configurable jobs to execute a backup and then push the resulting  backup to Amazon Glacier or S3. This aim is that frost can be easily configured to run various scripts as backups and can then be forgotten about except for receive email reports of the backups success/failure.

A "job" is a single command line command to execute resulting in one or more files that can be sent to Amazon Glacie or S3. Frosty takes care of setting environment variables and tidying up after itself to help ensure that no backups are left taking up disk space. The command that is run should produce one or more artifacts that will be zipped and sent to Amazon Glacier or S3.

Please note that Amazon Glacier and S3 are not equal. Please choose the service that is right for you ([this FAQ might help](https://aws.amazon.com/glacier/faqs/)). Notably, within Frosty, retention periods for backups are supported for S3 via lifecycles but with Glacier backups will remain indefinitely and must be archived externally for Frosty.  

# Usage

You need to create one or more backup scripts to be executed and then a configuration file detailing when to execute and where the resulting data should be stored Please see the following sections on how to do this.

## Commandline

```
Usage of frosty

	frosty <path-to-frosty-config-file> [flags...]

Flags:
  --validate
    	Validates that the specified config file is valid.
  --version
    	Prints the version information about the Frosty backup utility.
```

## Creating Backup Scripts

Frosty only accepts a single command with no arguments for each job. As such, it is recommended that you create shell scripts that Frosty will execute to run your backups. Any resulting artifacts from your script will be zipped up and pushed to the configured backup service.

In order to have backups pushed up to Amazon Glacier or S3 you must put any files you want backed up into the jobs `artifacts` directory. The location of this is available within your scripts from the environment variable `FROSTY_JOB_ARTIFACTS_DIR`. Please see the [examples](examples) directory for examples of how this is done.

**Note:** Make sure that your scripts have the correct permissions to run!

## Environment Variables

Frosty sets environment variables when running jobs for use within scripts called in the `command` property of a frosty job. The following environment variables are set by default:

- **FROSTY_JOB_DIR**: The absolute path to the working directory of the current job. This is of the form `~/.frosty/jobs/<job-name>`.
- **FROSTY_JOB_ARTIFACTS_DIR**: The absolute path to the folder that should contain any files that you want copied to Amazon Glacier or S3. This is of the form `~/.frosty/jobs/artefacts/<job-name>`.

## Configuration Files

The commands to execute to carry out a frosty backup are configured in a JSON file. By convention this should end with `.frosty.config` (e.g. `internal_server_backup.frosty.config`). A full example can be found in the [examples](examples) directory.

```javascript
// Main Config

{
  "reporting": {
    "email": {
      "smtp": {
        "host": "",     // String (required): The SMTP host name to connect to to send email reports.
        "port": "",     // String (required): The SMTP port number to connect to to send email reports.
        "username": "", // String (optional): The username for the SMTP account to connect to. If this is not provided not auth will be used.
        "password": ""  // String (optional): Must be supplied with username as the password for the SMTP account.
      },
      "sender": "",   // String (required): What sender address do you want on the email reports.
      "recipients": [ // String[] (required): A list of recipient email addresses that will get the reports.
        ""
      ]
    }
  },
  "backup": {
    // One of "s3" or "glacier" configuration. See below for more details.
  },
  "jobs": [ // Job[] (required): A list of configurations for jobs to be run.
    {
      "name": "",    // String (required): The name of the job to be run. This is how the job will be identified in the report.
      "command": "", // String (required): The shell command to run. This must not contain any arguments.
      "schedule": "" // String (required): Cron syntax for when the job should be scheduled.
    },
    ...
  ]
}


// s3 Config -- this should go in the "backup" property above if using S3.

"s3": {
  "accessKeyId": "",     // String (required): The AWS Access Key of the account you wish to use to store data to S3.
  "secretAccessKey": "", // String (required): The AWS Secret Key of the account you wish to use to store data to S3.
  "region": "",          // String (required): The AWS region you wish for your bucket to be created in. 
  "accountId": "",       // String (required): The AWS account ID you are using to store data in S3.
  "retentionDays":       // Int (optional): The number of days you wish to retain backups for. After this they will be automatically deleted.
}

 
// glacier Config -- this should go in the "backup" property above if using S3. 

"glacier": {
  "accessKeyId": "",     // String (required): The AWS Access Key of the account you wish to use to store data to S3.
  "secretAccessKey": "", // String (required): The AWS Secret Key of the account you wish to use to store data to S3.
  "region": "",          // String (required): The AWS region you wish for your bucket to be created in. 
  "accountId": ""        // String (required): The AWS account ID you are using to store data in S3.
}

 
```

# Reporting

## Emails

As detailed above in the "Configuration" section, frosty is able to send out glorious html email reports following each backup. Backups are grouped so that any jobs scheduled with the same time be batched together and result in a single report. That is, if all jobs have a cron of "0 1 * * *" all jobs will run at 1am and you will receive a single email shortly after this. If yoiu have two jobs with "0 1 * * *" and one job with "30 1 * * *" you will receive one email after the 1am jobs complete and 1 email after the 1:30am jobs complete.

If a job fails, it's standard out and standard error is added to the email so you can identify exactly what went wrong. The emails subject line will start with "[SUCCESS]" or "[FAILURE]" so you should be able to filter out success emails in your inbox if you are only interested in failures. 

A sample email report can be seen here:

![Frosty email report](https://i.imgur.com/GeW9Qek.png)
