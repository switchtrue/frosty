package reporting

import (
	"log"
	"net/smtp"
	"strings"
	"text/template"
)

type Mail struct {
	Host       string
	Port       string
	Sender     string
	Recipients []string
}

func (m *Mail) SetSMTPConnectionDetails(host string, port string) {
	m.Host = host
	m.Port = port
}

func (m *Mail) SetSender(sender string) {
	m.Sender = sender
}

func (m *Mail) AddRecipient(recipient string) {
	m.Recipients = append(m.Recipients, recipient)
}

func (m *Mail) GetSMTPHostAndPort() string {
	return m.Host + ":" + m.Port
}

func (m *Mail) GetRecipientHeader() string {
	var recipients string
	for _, recipient := range m.Recipients {
		recipients += recipient + ","
	}
	return strings.TrimRight(recipients, ",")
}

func (m *Mail) SendFromTemplate(tmpl *template.Template, templateData EmailSummaryTemplateData) {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial(m.GetSMTPHostAndPort())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Set the sender and recipient.
	c.Mail(m.Sender)
	c.Rcpt(m.Recipients[0])

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		log.Fatal(err)
	}
	defer wc.Close()

	wc.Write([]byte("From: " + m.Sender + "\n"))
	wc.Write([]byte("To: " + m.GetRecipientHeader() + "\n"))
	wc.Write([]byte("Subject: testing 1" + "\n"))
	wc.Write([]byte("Content-Type: text/html; charset=utf-8" + "\n"))
	wc.Write([]byte("\n"))

	tmpl.Execute(wc, templateData)
}
