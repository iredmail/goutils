package smtpclient

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

type Config struct {
	Host       string
	Port       string
	From       mail.Address
	Recipients []mail.Address
	Bcc        []mail.Address
	ReplyTo    mail.Address

	// smtp authentication
	SMTPUser     string
	SMTPPassword string

	StartTLS bool

	// mail subject / body
	Subject string
	Body    string
}

func Sendmail(c Config) error {
	if len(c.Recipients) == 0 {
		return errors.New("invalid recipients")
	}

	var toAddrs []string
	for _, addr := range c.Recipients {
		toAddrs = append(toAddrs, addr.String())
	}
	to := strings.Join(toAddrs, ",")

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = c.From.String()
	headers["To"] = to
	headers["Subject"] = c.Subject

	if len(c.Bcc) > 0 {
		var bccAddrs []string
		for _, addr := range c.Bcc {
			bccAddrs = append(bccAddrs, addr.String())
		}
		bcc := strings.Join(bccAddrs, ",")
		headers["Bcc"] = bcc
	}

	// FIXME 组装邮件的方式不严谨
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += c.Body

	client, err := smtp.Dial(net.JoinHostPort(c.Host, c.Port))
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", c.SMTPUser, c.SMTPPassword, c.Host)

	if c.StartTLS {
		tc := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         c.Host,
		}

		if err = client.StartTLS(tc); err != nil {
			return err
		}
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(c.From.Address); err != nil {
		return err
	}

	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	if _, err = w.Write([]byte(message)); err != nil {
		log.Fatalln(err)
	}

	err = w.Close()
	if err != nil {
		return err
	}

	if err = client.Quit(); err != nil {
		return err
	}

	return nil
}
