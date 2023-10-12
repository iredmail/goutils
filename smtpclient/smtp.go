package smtpclient

import (
	"crypto/tls"
	"errors"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/iredmail/goutils"
)

type Config struct {
	Host     string
	Port     string
	StartTLS bool
	Timeout  time.Duration

	// smtp authentication
	SMTPUser     string
	SMTPPassword string
}

func Sendmail(c Config, from string, to []string, subject string, body []byte, builder ...Builder) error {
	msg := message{
		from:    from,
		to:      to,
		subject: subject,
		body:    body,
	}

	for _, build := range builder {
		build(&msg)
	}

	encodeMsg, err := msg.Encode(c.Host)
	if err != nil {
		return err
	}

	if c.Timeout == 0 {
		c.Timeout = time.Second * 10
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.Host, c.Port), c.Timeout)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return err
	}

	defer func(client *smtp.Client) {
		_ = client.Quit()
	}(client)

	if err = client.Hello(goutils.GetHostFQDN()); err != nil {
		return err
	}

	hasExtSTARTTLS, _ := client.Extension("STARTTLS")
	if !hasExtSTARTTLS && c.StartTLS {
		return errors.New("smtp server unsupported tls")
	}

	if hasExtSTARTTLS && c.StartTLS {
		if err = client.StartTLS(&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         c.Host,
		}); err != nil {
			return err
		}
	}

	if len(c.SMTPUser) > 0 && len(c.SMTPPassword) > 0 {
		var auth smtp.Auth

		if ok, auths := client.Extension("AUTH"); ok {
			if strings.Contains(auths, "LOGIN") &&
				!strings.Contains(auths, "PLAIN") {
				auth = newLoginAuth(c.Host, c.SMTPUser, c.SMTPPassword)
			} else {
				auth = smtp.PlainAuth("", c.SMTPUser, c.SMTPPassword, c.Host)
			}
		}

		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	if _, err = w.Write(encodeMsg); err != nil {
		return err
	}

	return w.Close()
}

func SendmailInBackground(c Config, from string, to []string, subject string, body []byte, builder ...Builder) {
	go func() {
		// TODO return error
		_ = Sendmail(c, from, to, subject, body, builder...)
	}()
}

func SendmailWithEml(c Config, from mail.Address, to []string, emlPath string) error {
	// 读取邮件源码文件
	rawEmail, err := os.ReadFile(emlPath)
	if err != nil {
		return err
	}

	if c.Timeout == 0 {
		c.Timeout = time.Second * 10
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.Host, c.Port), c.Timeout)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return err
	}

	defer func(client *smtp.Client) {
		_ = client.Quit()
	}(client)

	if err = client.Hello(goutils.GetHostFQDN()); err != nil {
		return err
	}

	hasExtSTARTTLS, _ := client.Extension("STARTTLS")
	if !hasExtSTARTTLS && c.StartTLS {
		return errors.New("smtp server unsupported tls")
	}

	if hasExtSTARTTLS && c.StartTLS {
		if err = client.StartTLS(&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         c.Host,
		}); err != nil {
			return err
		}
	}

	if len(c.SMTPUser) > 0 && len(c.SMTPPassword) > 0 {
		var auth smtp.Auth

		if ok, auths := client.Extension("AUTH"); ok {
			if strings.Contains(auths, "LOGIN") &&
				!strings.Contains(auths, "PLAIN") {
				auth = newLoginAuth(c.Host, c.SMTPUser, c.SMTPPassword)
			} else {
				auth = smtp.PlainAuth("", c.SMTPUser, c.SMTPPassword, c.Host)
			}
		}

		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(from.Address); err != nil {
		return err
	}

	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	if _, err = w.Write(rawEmail); err != nil {
		return err
	}

	return w.Close()
}
