package mail

import (
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
)

type SMTPConn struct {
	smtpCfg SMTPConfig
	client  *smtp.Client
}

func (c *SMTPConn) Close() error {
	return c.client.Close()
}

func (c *SMTPConn) SendMail(e *Email) error {
	body, err := c.generateContent(e)
	if err != nil {
		return err
	}

	err = c.client.Mail(c.smtpCfg.User)
	if err != nil {
		return err
	}

	rcpts := append(e.ToList, e.CcList...)
	for _, r := range rcpts {
		err = c.client.Rcpt(r)
		if err != nil {
			return err
		}
	}

	w, err := c.client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write(body)
	if err != nil {
		return err
	}

	return nil
}

func (c *SMTPConn) generateContent(e *Email) ([]byte, error) {
	mm := email.NewEmail()
	mm.From = c.smtpCfg.User
	mm.To = e.ToList
	mm.Cc = e.CcList
	mm.Subject = e.Subject
	mm.Text = []byte(e.Body)
	data, err := mm.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate mail content: %s", err.Error())
	}
	return data, nil
}
