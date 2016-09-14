package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	SSL      bool
}

type MailerConfig struct {
	SMTP SMTPConfig
}

type Mailer struct {
	smtpCfg SMTPConfig
	tlsCfg  *tls.Config
	auth    smtp.Auth
}

func NewMailer(smtpCfg *SMTPConfig) (*Mailer, error) {
	var tlsConfig *tls.Config
	tlsConfig = &tls.Config{ServerName: smtpCfg.Host}
	auth := smtp.PlainAuth("", smtpCfg.User, smtpCfg.Password, smtpCfg.Host)
	m := &Mailer{
		smtpCfg: *smtpCfg,
		tlsCfg:  tlsConfig,
		auth:    auth,
	}
	return m, nil
}

type Email struct {
	ToList  []string
	CcList  []string
	Subject string
	Body    string
}

func (m *Mailer) NewSMTPConnection() (*SMTPConn, error) {
	var client *smtp.Client
	var err error
	if m.smtpCfg.SSL {
		client, err = m.connectWithSSL()
	} else {
		client, err = m.connect()
	}
	if err != nil {
		return nil, err
	}

	conn := &SMTPConn{
		smtpCfg: m.smtpCfg,
		client:  client,
	}

	return conn, nil
}

func (m *Mailer) connectWithSSL() (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", m.smtpCfg.Host, m.smtpCfg.Port)
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: time.Second * 30}, "tcp", addr, m.tlsCfg)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		return nil, err
	}

	client, err := smtp.NewClient(conn, m.smtpCfg.Host)
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = client.Auth(m.auth)
	if err != nil {
		return nil, err
	}

	conn = nil
	return client, err
}

func (m *Mailer) connect() (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", m.smtpCfg.Host, m.smtpCfg.Port)
	conn, err := net.DialTimeout("tcp", addr, time.Second*30)
	if err != nil {
		return nil, err
	}

	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	client, err := smtp.NewClient(conn, m.smtpCfg.Host)
	if err != nil {
		return nil, err
	}

	if ok, _ := client.Extension("StartTLS"); ok {
		err = client.StartTLS(m.tlsCfg)
		if err != nil {
			return nil, err
		}
	}

	err = client.Auth(m.auth)
	if err != nil {
		return nil, err
	}

	conn = nil
	return client, err
}
