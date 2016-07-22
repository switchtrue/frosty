package reporting

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"text/template"
)

type Mail struct {
	Host       string
	Port       string
	Username   string
	Password   string
	Sender     string
	Recipients []string
}

func (m *Mail) AddRecipient(recipient string) {
	m.Recipients = append(m.Recipients, recipient)
}

func (m *Mail) getSMTPHostAndPort() string {
	return m.Host + ":" + m.Port
}

func (m *Mail) recipientHeader() string {
	var recipients string
	for _, recipient := range m.Recipients {
		recipients += recipient + ","
	}
	return strings.TrimRight(recipients, ",")
}

func (m *Mail) useAuthentication() bool {
	return m.Username != ""
}

func (m *Mail) SendFromTemplate(tmpl *template.Template, templateData EmailSummaryTemplateData) {
	var subject string
	if templateData.IsSuccessful() {
		subject = "[SUCCESS] Frosty Backup Report"
	} else {
		subject = "[FAILURE] Frosty Backup Report"
	}

	var wc bytes.Buffer

	headers := make(map[string]string)
	headers["From"] = m.Sender
	headers["To"] = m.recipientHeader()
	headers["Subject"] = subject
	headers["Content-Type"] = "text/html; charset=utf-8"

	for k, v := range headers {
		h := fmt.Sprintf("%s: %s\r\n", k, v)
		wc.Write([]byte(h))
	}

	wc.Write([]byte("\r\n"))

	tmpl.Execute(&wc, templateData)

	if m.useAuthentication() {
		a := smtp.PlainAuth("", m.Username, m.Password, m.Host)
		err := smtp.SendMail(m.getSMTPHostAndPort(), a, m.Sender, m.Recipients, wc.Bytes())
		if err != nil {
			log.Fatalf("Error sending email with auth: %s\n", err)
		}
	} else {
		// Connect to the remote SMTP server.
		c, err := smtp.Dial(m.getSMTPHostAndPort())
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()

		// Set the sender and recipient.
		c.Mail(m.Sender)
		c.Rcpt(m.Recipients[0])

		// Send the email body.
		cwc, err := c.Data()
		if err != nil {
			log.Fatal(err)
		}
		defer cwc.Close()

		cwc.Write(wc.Bytes())
	}
}
