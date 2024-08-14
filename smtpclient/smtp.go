package smtpclient

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/iredmail/goutils"
	"github.com/iredmail/goutils/logger"
)

type Config struct {
	Host    string
	Port    string
	Timeout time.Duration

	StartTLS             bool
	UseSSL               bool
	VerifySSLCertificate bool

	// smtp authentication
	DisplayName  string
	SMTPUser     string
	SMTPPassword string
}

func SendmailWithComposer(c Config, composer *Composer) (err error) {
	// 如果 from 不是完整邮件地址，则将 smtp 主机名作为邮件地址的域名部分追加到 from 拼凑成完整邮件地址。
	// 例如：user@domain.com、 user@[IP]
	if len(composer.from.Address) > 0 && !strings.Contains(composer.from.Address, "@") {
		if goutils.IsIP(c.Host) {
			composer.from.Address = fmt.Sprintf("%s@[%s]", composer.from.Address, c.Host)
		} else {
			composer.from.Address = fmt.Sprintf("%s@%s", composer.from.Address, c.Host)
		}
	}

	// 如果 composer.From 为空并且 SMTPUser 不为空，那么将 SMTPUser 赋值 composer.From 字段
	if composer.from.Address == "" && c.SMTPUser != "" {
		if goutils.IsIP(c.Host) {
			composer.from.Address = fmt.Sprintf("%s@[%s]", c.SMTPUser, c.Host)
		} else {
			composer.from.Address = fmt.Sprintf("%s@%s", c.SMTPUser, c.Host)
		}
	}

	if composer.from.Address == "" {
		composer.from.Address = fmt.Sprintf("%s:%s", c.Host, c.Port)
	}

	// Export mail body before smtp connection, make sure it's valid email message.
	msg, err := composer.Bytes()
	if err != nil {
		return fmt.Errorf("failed in building email message from composer: %v", err)
	}

	if c.Timeout == 0 {
		c.Timeout = time.Second * 15
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.Host, c.Port), c.Timeout)
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

	if err = client.Mail(composer.from.Address); err != nil {
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

/*
	func SendmailWithEml(c Config, from mail.Address, recipients []string, emlPath string) (err error) {
	smtpServer := fmt.Sprintf("%s:%s", c.Host, c.Port)

	client, err := smtp.Dial(smtpServer)
	if err != nil {
		return err
	}

	defer func() {
		_ = client.Close()
	}()

	// HELO
	domain := emailutils.ExtractDomain(from.Address)
	if domain == "" {
		domain = "example.com"
	}

	err = client.Hello(domain)
	if err != nil {
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

	// AUTH
	auth := smtp.PlainAuth("", c.SMTPUser, c.SMTPPassword, c.Host)
	err = client.Auth(auth)
	if err != nil {
		return err
	}

	// MAIL
	err = client.Mail(from.Address)
	if err != nil {
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
