package smtpclient

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/iredmail/goutils/emailutils"
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

	if err = client.Hello(emailutils.ExtractDomain(from)); err != nil {
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
			if strings.Contains(auths, "CRAM-MD5") {
				auth = smtp.CRAMMD5Auth(c.SMTPUser, c.SMTPPassword)
			} else if strings.Contains(auths, "LOGIN") &&
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
		log.Fatalln(err)
	}

	if err = w.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func SendmailInBackground(c Config, from string, to []string, subject string, body []byte, builder ...Builder) {
	go func() {
		// TODO return error
		_ = Sendmail(c, from, to, subject, body, builder...)
	}()
}

func SendmailWithEml(c Config, from mail.Address, recipients []string, emlPath string) error {
	smtpServer := fmt.Sprintf("%s:%s", c.Host, c.Port)

	// CONNECT
	client, err := smtp.Dial(smtpServer)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	// HELO
	domain := emailutils.ExtractDomain(from.Address)
	if domain == "" {
		domain = "example.com"
	}
	if err = client.Hello(domain); err != nil {
		return err
	}

	// TLS
	if c.StartTLS {
		tlsConfig := &tls.Config{
			// ServerName:         utils.GetHostName(),
			InsecureSkipVerify: true,
		}
		err = client.StartTLS(tlsConfig)
		if err != nil {
			fmt.Printf("failed in STARTTLS directive: %v\n", err)
			os.Exit(255)
		}
	}

	auth := smtp.PlainAuth("", c.SMTPUser, c.SMTPPassword, c.Host)
	if err = client.Auth(auth); err != nil {
		return err
	}

	// `MAIL FROM:`
	if err = client.Mail(from.Address); err != nil {
		return err
	}

	// `RCPT TO:`
	var toAddrs []string
	for _, addr := range recipients {
		toAddrs = append(toAddrs, addr)
	}
	to := strings.Join(toAddrs, ",")
	if err = client.Rcpt(to); err != nil {
		return err
	}

	// `DATA`
	w, err := client.Data()
	if err != nil {
		return err
	}

	// 读取邮件源码文件
	rawEmail, err := os.ReadFile(emlPath)
	if err != nil {
		return err
	}

	_, err = w.Write(rawEmail)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
