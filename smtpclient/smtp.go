package smtpclient

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"

	"github.com/jhillyerd/enmime"
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

	return client.Quit()
}

func SendmailWithEml(c Config, emlPath string) error {
	emlBytes, err := os.ReadFile(emlPath)
	if err != nil {
		return err
	}

	headers, err := getEmlHeaders(emlBytes)
	if err != nil {
		return err
	}

	body, err := getEmlBody(emlBytes)
	if err != nil {
		return err
	}

	if len(c.Recipients) == 0 {
		return errors.New("invalid recipients")
	}

	var toAddrs []string
	for _, addr := range c.Recipients {
		toAddrs = append(toAddrs, addr.String())
	}
	to := strings.Join(toAddrs, ",")

	// reset headers
	headers["From"] = c.From.String()
	headers["To"] = to

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
	message += "\n\n" + string(body)

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

	return client.Quit()
}

func getEmlHeaders(emailBytes []byte) (headers map[string]string, err error) {
	headers = make(map[string]string)
	envelope, err := enmime.NewParser().ReadEnvelope(bytes.NewReader(emailBytes))
	if err != nil {
		return nil, err
	}

	for _, key := range envelope.GetHeaderKeys() {
		headers[key] = envelope.GetHeader(key)
	}

	return
}

func getEmlBody(emailBytes []byte) (body []byte, err error) {
	headersLength := 0
	br := bufio.NewReader(bytes.NewReader(emailBytes))
	for {
		// Pull out each line of the headers as a temporary slice s
		s, err := br.ReadSlice('\n')
		if err != nil {
			return nil, err
		}

		firstColon := bytes.IndexByte(s, ':')
		firstSpace := bytes.IndexAny(s, " \t\n\r")

		if firstSpace == 0 {
			headersLength += len(s)

			continue
		}

		if firstColon < 0 {
			break
		} else {
			headersLength += len(s)
		}
	}

	body = emailBytes[headersLength:]

	return
}
