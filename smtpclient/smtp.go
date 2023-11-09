package smtpclient

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/iredmail/goutils"
	"github.com/iredmail/goutils/emailutils"
	"github.com/iredmail/goutils/logger"
)

type Config struct {
	Host string
	Port string

	StartTLS bool
	StartSSL bool

	// Sender
	From mail.Address

	// smtp authentication
	SMTPUser     string
	SMTPPassword string

	timeout time.Duration
}

func SendmailWithComposer(c Config, composer *Composer) (err error) {
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
	if c.StartSSL {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
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
			InsecureSkipVerify: true,
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

	if err = client.Mail(c.From.Address); err != nil {
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

// Sendmail 发送邮件。
// 简化邮件格式。
func Sendmail(c Config, recipients, bcc []string, replyTo, subject, body string) error {
	// 过滤出有效的邮件地址
	var rcptAddrs []string
	for _, rcpt := range recipients {
		if emailutils.IsEmail(rcpt) {
			rcptAddrs = append(rcptAddrs, rcpt)
		}
	}

	if len(rcptAddrs) == 0 {
		return errors.New("invalid recipients")
	}

	// 构造邮件头
	headers := map[string]string{
		"From":       c.From.String(),
		"To":         strings.Join(rcptAddrs, ","),
		"Subject":    subject,
		"Message-ID": fmt.Sprintf("<%s@%s>", goutils.GenRandomString(32), c.Host),
		"Date":       time.Now().UTC().Format(time.RFC1123Z),
	}

	if emailutils.IsEmail(replyTo) {
		headers["Reply-To"] = replyTo
	}

	if len(bcc) > 0 {
		var bccAddrs []string

		for _, addr := range bcc {
			bccAddrs = append(bccAddrs, addr)
		}

		headers["Bcc"] = strings.Join(bccAddrs, ",")
	}

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	// 邮件 header 和 body 以第一个空白行作为分界
	message += "\r\n" + body

	client, err := smtp.Dial(net.JoinHostPort(c.Host, c.Port))
	if err != nil {
		return err
	}

	err = client.Hello(goutils.GetHostFQDN())
	if err != nil {
		return err
	}

	if c.StartTLS {
		tc := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         c.Host,
		}

		if err = client.StartTLS(tc); err != nil {
			return err
		}

		err = client.Hello(goutils.GetHostFQDN())
		if err != nil {
			return err
		}
	}

	if len(c.SMTPUser) > 0 && len(c.SMTPPassword) > 0 {
		auth := smtp.PlainAuth("", c.SMTPUser, c.SMTPPassword, c.Host)

		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(c.From.Address); err != nil {
		return err
	}

	for _, rcpt := range rcptAddrs {
		if err = client.Rcpt(rcpt); err != nil {
			return err
		}
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

	return client.Quit()
}

// SendmailInBackground 在后台发送邮件，不阻塞当前进程。
/*
func SendmailInBackground(c Config, recipients, bcc []string, replyTo, subject, body string, l logger.Logger) {
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

		err := Sendmail(c, recipients, bcc, replyTo, subject, body)
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
*/
