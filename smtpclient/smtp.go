package smtpclient

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/iredmail/goutils"
	"github.com/iredmail/goutils/logger"
)

type Config struct {
	Host string
	Port string

	StartTLS             bool
	UseSSL               bool
	VerifySSLCertificate bool

	// Sender
	From mail.Address

	// smtp authentication
	SMTPUser     string
	SMTPPassword string

	timeout time.Duration
}

func SendmailWithComposer(c Config, composer *Composer) (err error) {
	// 如果 from 不是一个邮件地址，那么需要将 from 和 host 进行组装
	// - Domain: from@domain
	// - IP: from@[IP]
	if !strings.Contains(composer.from.Address, "@") {
		if goutils.IsIPv4(c.Host) {
			composer.from.Address = fmt.Sprintf("%s@[%s]", composer.from.Address, c.Host)
		} else {
			composer.from.Address = fmt.Sprintf("%s@%s", composer.from.Address, c.Host)
		}
	}

	// Export mail body before smtp connection, make sure it's valid email message.
	msg, err := composer.Bytes()
	if err != nil {
		return fmt.Errorf("failed in building email message from composer: %v", err)
	}

	if c.timeout == 0 {
		c.timeout = time.Second * 15
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.Host, c.Port), c.timeout)
	if err != nil {
		return err
	}

	// 开启 ssl 安全连接
	if c.UseSSL {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: !c.VerifySSLCertificate,
			ServerName:         c.Host,
		}

		conn = tls.Server(conn, tlsConfig)
	}

	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return err
	}

	if c.StartTLS {
		tc := &tls.Config{
			InsecureSkipVerify: !c.VerifySSLCertificate,
			ServerName:         c.Host,
		}

		if err = client.StartTLS(tc); err != nil {
			return err
		}
	}

	if len(c.SMTPUser) > 0 && len(c.SMTPPassword) > 0 {
		auth := smtp.PlainAuth("", c.SMTPUser, c.SMTPPassword, c.Host)

		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(c.SMTPUser); err != nil {
		return err
	}

	for _, addr := range composer.GetTo() {
		if err = client.Rcpt(addr.Address); err != nil {
			return err
		}
	}

	for _, addr := range composer.cc {
		if err = client.Rcpt(addr.Address); err != nil {
			return err
		}
	}

	for _, addr := range composer.bcc {
		if err = client.Rcpt(addr.Address); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	if _, err = w.Write(msg); err != nil {
		log.Fatalln(err)
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// SendmailWithComposerInBackground 在后台发送邮件，不阻塞当前进程。
func SendmailWithComposerInBackground(c Config, composer *Composer, l logger.Logger) {
	go func() {
		// 捕捉 panic 并记录具体信息，便于后期排错。
		defer func() {
			if r := recover(); r != nil {
				if l != nil {
					l.Error("panic in SendmailWithComposer: %v\n%s", r, debug.Stack())
				} else {
					fmt.Printf("panic in SendmailWithComposer: %v\n%s", r, debug.Stack())
				}
			}
		}()

		err := SendmailWithComposer(c, composer)
		if err != nil {
			if l != nil {
				l.Error("Failed in sending email: %v", err)
			} else {
				fmt.Printf("Failed in sending email: %v\n", err)
			}
		} else {
			if l != nil {
				var addrs []string
				for _, addr := range composer.GetTo() {
					addrs = append(addrs, addr.Address)
				}

				l.Info("Email sent. Subject='%s', To='%s'", composer.GetSubject(), strings.Join(addrs, ","))
			}
		}
	}()
}
