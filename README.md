# !!NOTE!!

This currently is not finished!

# Frosty Backup Utility

A lightweight command line backup utility that stores back ups as archives in Amazon Glacier.

Frosty will run one or more user configurable jobs to execute a backup and then push the resulting  backup to Amazon Glacier taking advantage of its low storage costs. The idea is to keep a single configuration file detailing the various system backups that might be required, executing them all via single command and receiving a single email report on success/failure.

A "job" is a single command line command to execute resulting in one or more files that can be sent to Amazon Glacier. Frosty takes care of setting environment variables and tidying up after itself to help ensure that no backups are left taking up disk space.

The command that is run should produce one or more artefacts that will be zipped and sent to Amazon Glacier.

# Usage

## Configuration

The commands to execute to carry out a frosty backup are configured in a JSON file. By convention this should end with `.frosty.config` (e.g. `internal_server_backup.frosty.config`).

### .frosty.config

This JSON file can contain the following properties (see below for an example file):

- **jobs** *(mandatory)*: This contains a list of named jobs that frosty will run. Each "job" should be a discreet task to run to backup another program. Each job will be timed and marked as a success or failure in the report.
- **jobs.name** *(mandatory)*: The name of this job. This will be included in any reporting and must be unique throughout the config file.
- **jobs.command** *(mandatory)*: The commands to execute to run this job. This can only be a single command with no arguments - if you require a non-trivial backup script it is recommended that you save this in a separate file and invoke that file from the `command` property (e.g `"command": ["./path/to/my/script.sh"]`).
- **reporting**: This will contain information for optionally configuring how reports of each backup managed. For now, only email reports are supported.
- **reporting.email**: This contains the configuration for email reports.
- **reporting.email.smtp.host** *(mandatory)*: The host name of the SMTP server to use.
- **reporting.email.smtp.port** *(mandatory)*: The port of the SMTP server to use.
- **reporting.email.sender** *(mandatory)*: The email address of the sender or the email (e.g. frosty-noreply@email.com)
- **reporting.email.recipients** *(mandatory)*: A JSON list of recipients of the email report.

### Shell Scripts

Frosty only accepts a single command with no arguments for each job. As such, it is recommended that you create shell scripts that Frosty will execute to run your backups.

In order to have backups pushed up to Amazon Glacier you must put any files you want backed up into the jobs `artefacts` directory. The location of this is available within your scripts from the environment variable `FROSTY_JOB_ARTIFACTS_DIR`. Please see the [examples](examples) directory for examples of how this is done.

**Note:** Make sure that your scripts have the correct permissions to run!

### Environment Variables

Frosty sets environment variables when running jobs for use within scripts called in the `command` property of a frosty job. The following environment variables are set by default:

- **FROSTY_JOB_DIR**: The absolute path to the working directory of the current job. This is of the form `~/.frosty/jobs/<job-name>`.
- **FROSTY_JOB_ARTIFACTS_DIR**: The absolute path to the folder that should contain any files that you want copied to Amazon Glacier. This is of the form `~/.frosty/jobs/artefacts/<job-name>`.

### Example .frosty.config file

```json
{
  "reporting": {
    "email": {
      "smtp": {
        "host": "smtp.domain.com",
        "port": 25
      },
      "sender": "frosty-noreply@email.com",
      "to": [
        "foo@bar.com",
        "joe.bloggs@email.com"
      ]
    }
  },
  "jobs": [
    {
      "name": "backup-something",
      "command": "scripts/backup_something.sh"
    },
    {
      "name": "backup-something-else",
      "command": "scripts/backup_something_else.sh"
    },
  ]
}
```

## Starting a backup
`frosty backup /path/to/my.frosty.config`

## Validating your .frosty.config file
`frosty validate /path/to/my.frosty.config`

## Running Nightly Backups
To run nightly backups it is recommended that you set up a [cron](https://en.wikipedia.org/wiki/Cron) job to execute `frosty backup /path/to/my.frosty.config`.

# Reporting

## Emails

As detailed above in the "Configuration" section, frosty is able to send out glorious html email reports following each backup. A sample email report can be see here:

![Frosty email report](https://i.imgur.com/GeW9Qek.png)

If a job fails, it's standard out is added to the email so you can identify exactly what went wrong.



