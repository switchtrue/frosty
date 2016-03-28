package reporting

import (
	"log"
	"net/smtp"
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

func (m *Mail) SendFromTemplate(tmpl *template.Template, templateData EmailSummaryTemplateData) {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial(m.GetSMTPHostAndPort())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Set the sender and recipient.
	c.Mail("sender@example.org")
	c.Rcpt("recipient@example.net")

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		log.Fatal(err)
	}
	defer wc.Close()

	wc.Write([]byte("From: " + m.Sender + "\n"))
	wc.Write([]byte("To: " + m.Recipients[0] + "\n"))
	wc.Write([]byte("Subject: testing 1" + "\n"))
	wc.Write([]byte("Content-Type: text/html; charset=utf-8" + "\n"))
	wc.Write([]byte("\n"))

	tmpl.Execute(wc, templateData)
}
