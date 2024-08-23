package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendEmailUsingGmail(t *testing.T) {

	if testing.Short() {
		t.Skip()
	}

	c, err := LoadConfig("../../")
	require.NoError(t, err)

	g := NewGmailSender(c.EmailSenderName, c.GmailSenderAddress, c.GmailSenderPassword)

	subject := "Test email"
	content := `
	<h2>Hello from go</h2>
	<p>This is a test message from util.email_test.go.</p>
	`
	to := []string{"guerzon@proton.me"}
	files := []string{"../../README.md"}

	err = g.SendEmail(subject, content, to, nil, nil, files)
	require.NoError(t, err)
}

func TestSendEmailUsingMailhog(t *testing.T) {

	if testing.Short() {
		t.Skip()
	}

	c, err := LoadConfig("../../")
	require.NoError(t, err)

	m := NewLocalSender(c.EmailSenderName, c.MailhogSenderAddress, c.MailhogHost)

	subject := "Test email"
	content := `
	<h2>Hello from go</h2>
	<p>This is a test message from util.email_test.go.</p>
	`
	to := []string{"guerzon@proton.me"}

	err = m.SendEmail(subject, content, to, nil, nil, nil)
	require.NoError(t, err)
}
