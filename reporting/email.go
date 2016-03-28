package reporting

import (
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/mleonard87/frosty/config"
	"github.com/mleonard87/frosty/job"
)

type EmailSummaryTemplateData struct {
	StartTime   time.Time
	EndTime     time.Time
	ElapsedTime time.Duration
	Hostname    string
	Jobs        []job.JobStatus
	Status      int
}

func (estd EmailSummaryTemplateData) IsSuccessful() bool {
	return estd.Status == job.STATUS_SUCCESS
}

func SendEmailSummary(jobStatuses []job.JobStatus, emailConfig *config.EmailReportingConfig) {

	templateData := getEmailSummaryTemplateData(jobStatuses)

	t, _ := template.ParseFiles("tmpl/email_summary.html")

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
		fmt.Println("Could not determine hostname.")
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

	return EmailSummaryTemplateData{
		startTime,
		endTime,
		endTime.Sub(startTime),
		hostname,
		jobStatuses,
		status,
	}
}
