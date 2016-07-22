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
	//// Connect to the remote SMTP server.
	//c, err := smtp.Dial(m.GetSMTPHostAndPort())
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer c.Close()
	//
	//// Set the sender and recipient.
	//c.Mail(m.Sender)
	//c.Rcpt(m.Recipients[0])
	//
	////c.Auth()
	//
	//// Send the email body.
	//wc, err := c.Data()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer wc.Close()
	//
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
		fmt.Println("Sending email with auth")
		a := smtp.PlainAuth("", m.Username, m.Password, m.Host)
		err := smtp.SendMail(m.getSMTPHostAndPort(), a, m.Sender, m.Recipients, wc.Bytes())
		if err != nil {
			log.Printf("Error sending email with auth: %s\n", err)
		}
	} else {
		fmt.Println("Sending email without auth")
		err := smtp.SendMail(m.getSMTPHostAndPort(), nil, m.Sender, m.Recipients, wc.Bytes())
		if err != nil {
			log.Printf("Error sending email without auth: %s\n", err)
		}
	}
}
