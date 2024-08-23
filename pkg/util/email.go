package util

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	gmailAuthHost       = "smtp.gmail.com"
	gmailSubmissionHost = "smtp.gmail.com:587"
)

type EmailSender interface {
	SendEmail(subject string, content string, to []string, cc []string, bcc []string, attachFiles []string) error
}

type GmailSender struct {
	name           string
	sender         string
	password       string
	authHost       string
	submissionHost string
}

type LocalSender struct {
	name       string
	sender     string
	smtpServer string
}

// Creates a new Gmail sender
func NewGmailSender(senderName string, senderAddress string, senderPassword string) EmailSender {
	return &GmailSender{
		name:           senderName,
		sender:         senderAddress,
		password:       senderPassword,
		authHost:       gmailAuthHost,
		submissionHost: gmailSubmissionHost,
	}
}

// Creates a new local email sender
func NewLocalSender(senderName string, senderAddress string, smtpServer string) EmailSender {
	return &LocalSender{
		name:       senderName,
		sender:     senderAddress,
		smtpServer: smtpServer,
	}
}

// Sends an email using a Gmail sender address
func (g *GmailSender) SendEmail(subject string, content string, to []string, cc []string, bcc []string, attachFiles []string) error {

	email := email.NewEmail()
	email.From = fmt.Sprintf("%s <%s>", g.name, g.sender)
	email.Subject = subject
	email.HTML = []byte(content)
	email.To = to
	email.Cc = cc
	email.Bcc = bcc

	for _, v := range attachFiles {
		_, err := email.AttachFile(v)
		if err != nil {
			return fmt.Errorf("cannot attach file %s: %w", v, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", g.sender, g.password, g.authHost)

	return email.Send(g.submissionHost, smtpAuth)
}

// Sends and email using a local sender to an unauthenticated SMTP server such as MailHog.
func (l *LocalSender) SendEmail(subject string, content string, to []string, cc []string, bcc []string, attachFiles []string) error {

	email := email.NewEmail()
	email.From = fmt.Sprintf("%s <%s>", l.name, l.sender)
	email.Subject = subject
	email.HTML = []byte(content)
	email.To = to

	return email.Send(l.smtpServer, nil)
}
