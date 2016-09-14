package mail

import (
	"testing"
)

func newTestConfig() *SMTPConfig {
	cfg := &SMTPConfig{}
	cfg.Host = "smtp.exmail.qq.com"
	cfg.Port = 25
	cfg.User = "mozhonghua@sursen.net"
	cfg.Password = "Cv$xWKHM/5qS1Zd5"
	cfg.SSL = false
	return cfg
}

func newTestEmail() *Email {
	e := &Email{}
	e.Subject = "当前时段统计报表"
	e.Body = "test email body"
	e.ToList = append(e.ToList, "mozhonghua@sursen.net")
	e.CcList = append(e.CcList, "saritald@163.com")

	return e
}

func testSendMail(t *testing.T, m *Mailer) {
	conn, err := m.NewSMTPConnection()
	if err != nil {
		t.Fatal(err)
		return
	}
	defer conn.Close()

	e := newTestEmail()
	if err := conn.SendMail(e); err != nil {
		t.Fatal(err)
		return
	}
}

func TestSendMailStartTLS(t *testing.T) {
	cfg := newTestConfig()
	m, err := NewMailer(cfg)
	if err != nil {
		t.Fatal(err)
		return
	}
	testSendMail(t, m)
}

func TestSendMailWithSSL(t *testing.T) {
	cfg := newTestConfig()
	cfg.Port = 465
	cfg.SSL = true

	m, err := NewMailer(cfg)
	if err != nil {
		t.Fatal(err)
		return
	}
	testSendMail(t, m)
}
