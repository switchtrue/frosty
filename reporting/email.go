package reporting

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/mleonard87/frosty/backup"
	"github.com/mleonard87/frosty/config"
	"github.com/mleonard87/frosty/job"
	"github.com/mleonard87/frosty/tmpl"
)

type EmailSummaryTemplateData struct {
	StartTime   time.Time
	EndTime     time.Time
	ElapsedTime time.Duration
	Hostname    string
	VaultName   string
	Jobs        []job.JobStatus
	Status      int
}

func (estd EmailSummaryTemplateData) IsSuccessful() bool {
	return estd.Status == job.STATUS_SUCCESS
}

func SendEmailSummary(jobStatuses []job.JobStatus, emailConfig *config.EmailReportingConfig) {
	templateData := getEmailSummaryTemplateData(jobStatuses)

	data, err := tmpl.Asset("tmpl/email_summary.html")
	if err != nil {
		fmt.Println(err)
	}

	t := template.New("frosty-report")
	t, err2 := t.Parse(string(data))
	if err2 != nil {
		fmt.Println(err2)
	}

	mail := Mail{}

	mail.SetSMTPConnectionDetails(emailConfig.SMTP.Host, emailConfig.SMTP.Port)
	mail.SetSender(emailConfig.Sender)
	for _, recipient := range emailConfig.Recipients {
		mail.AddRecipient(recipient)
	}

	mail.SendFromTemplate(t, templateData)
}

func getEmailSummaryTemplateData(jobStatuses []job.JobStatus) EmailSummaryTemplateData {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Could not determine hostname.", err)
	}

	var startTime, endTime time.Time
	status := job.STATUS_SUCCESS

	for _, j := range jobStatuses {
		if startTime.IsZero() || j.StartTime.Before(startTime) {
			startTime = j.StartTime
		}

		if endTime.IsZero() || j.EndTime.After(endTime) {
			endTime = j.EndTime
		}

		if j.Status == job.STATUS_FAILURE {
			status = job.STATUS_FAILURE
		}
	}

	vaultName := backupservice.GetBackupName()

	return EmailSummaryTemplateData{
		startTime,
		endTime,
		endTime.Sub(startTime),
		hostname,
		vaultName,
		jobStatuses,
		status,
	}
}
