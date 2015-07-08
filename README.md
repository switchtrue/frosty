# Disclaimer!

There is currently no code for this project! The readme file here is intended initially as a set of requirements for myself! However, pklease keep reading to discover its intent and watch this repository to see updates!

# frosty
A lightweight command line backup utility that stores back ups as archives in Amazon Glacier.

Frosty will run one or more user configurable jobs to execute a backup and then push the resulting  backup to Amazon Glacier taking advantage of its low storage costs. The idea is to keep a single configuration file detailing the various system backups that might be required, executing them all via single command and receiving a single email report on success/failure.

A "job" is a series of command line tasks to execute resulting in one or more files that can be sent to Amazon Glacier. Frosty takes care of setting environment variables and tidying up after itself to help ensure that now backups are left taking up disk space on the server in which they were run.

# Usage

## Configuration

The commands to execute to carry out a frosty backup are configured in a JSON file. By convention this should end with `.frosty.config` (e.g. `internal_server_backup.frosty.config`).

### .frosty.config

This JSON file can contain the following properties (see below for an example file):

- **frostyHome** *(mandatory)*: The working directory that frosty will use to create to deal with temporary files created during the backup process. This folder must exist and the frost utility must have read, write and execute permissions on it. As each job runs the value of this will also be set in the $FROSTY_HOME environment variable. It is recommended that any backup files created ask part of a job are placed within this directory (see the Environment Variables section below).
- **reportRecipients**: A list of email addresses that will receive an email containing a report of the backup jobs everytime it is run.
- **errorReportRecipients**: A list of email addresses that will receive an email containing a report of the backup jobs if a failure has occurred. The error message(s) will be attached to the email. Whilst this property is optional, if it is not set backups will fail silently.
- **envVars**: This optional property can contain key-value pairs of environment variable names and their values which will be set for all jobs.
- **jobs** *(mandatory)*: This contains a list of named jobs that frosty will run. Each "job" should be a discreet task to run to backup another program. Each job will be timed and marked as a success or failure in the report.
- **jobs.*.envVars**: This optional property can contain key-value pairs of environment variable names and their values which will be set for this job only.
- **jobs.*.script** *(mandatory)*: The commands to execute to run this job. If you have a particularly complicated backup script it is recommended that you save this in a separate file and invoke that file from the `script` property (e.g `"script": ["./path/to/my/script.sh"]`).
- **jobs.*.archives** *(mandatory)*: A list of file system [globs](https://en.wikipedia.org/wiki/Glob_(programming)). each matched file for this glob will be pushed to Amazon Glacier.

### Environment Variables

Frosty sets environment variables when running jobs for use within scripts that may be called or in the `script` property of a frosty job. The following environment variables are set by default:

- **FROSTY_HOME**: This is set to the value of the `frostyHome` property in the .frosty.config file. This is where frosty will carry out all of its backups and write its own temporary files. This will not contain a trailing slash even is one is specified in the `frostyHome` property.
- **FROSTY_RUN_HOME**: This will be set to the path of a directory created within `FROSTY_HOME` that will be created for the specific run of frosty. The name of this directory will be the current date and time (e.g. `$FROSTY_HOME/20150721_1836`). This will not contain a trailing slash. This directory will be deleted once frosty has completed the backup regardless of success or failure.
- **FROSTY_JOB_HOME**: This will be set to the path of a directory created within `FROSTY_RUN_HOME` that will be created for the specific job running. The name of this directory will be the name of the job (e.g. `$FROSTY_RUN_HOME/postgres`). This directory will be deleted once the entire backup has completed regardless of success or failure.

### Example .frosty.config file

```json
{
  "frostyHome": "",

  "reportRecipients": [
    "techteam@mycompany.com",
    "boss@mycompany.com"
  ],

  "errorReportRecipients": [
    "admin@mycompany.com",
    "techteam@mycompany.com"
  ],

  "envVars": {
    "hello": "mike",
  },

  "jobs": {
    "svn": {
      "script": [
        "mkdir $FROSTY_WORKING_DIRECTORY/svn_backup",
        "svnadmin hotcopy /var/www/svn/SEWPaC_POC $FROSTY_WORKING_DIRECTORY/svn_backup --clean-logs",
        "cd $FROSTY_WORKING_DIRECTORY && zip -r svn_backup.zip svn_backup"
      ],
      "archives": ["$FROSTY_WORKING_COPY/svn_backup.zip"]
    },
    "postgres": {
      "envVars": {
        "PGPASSWORD": "p@ssw0rd"
      },
      "script": [
        "pg_dump dbname -U username | gzip > $FROSTY_WORKING_DIRECTORY/postgres_backup.gz"
      ],
      "archives": "$FROSTY_WORKING_DIRECTORY/postgres_backup.gz"
    },
  }
}
```

## Starting a backup
`frosty backup /path/to/my.frosty.config`

## Running Nightly Backups
To run nightly backups it is recommended that you set up a [cron](https://en.wikipedia.org/wiki/Cron) job to execute `frosty backup /path/to/my.frosty.config`.
