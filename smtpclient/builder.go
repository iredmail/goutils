package smtpclient

import (
	"bytes"
	"fmt"
	"net/mail"
	"time"

	"github.com/jhillyerd/enmime"

	"github.com/iredmail/goutils"
)

type Builder func(b *message)

func WithCc(cc []string) Builder {
	return func(b *message) {
		b.cc = cc
	}
}

func WithBcc(bcc []string) Builder {
	return func(b *message) {
		b.bcc = bcc
	}
}

func WithReplyTo(replies []string) Builder {
	return func(b *message) {
		b.replyTo = replies
	}
}

func WithHtml(html []byte) Builder {
	return func(b *message) {
		b.html = html
	}
}

func WithHeaders(headers map[string]string) Builder {
	return func(b *message) {
		b.headers = headers
	}
}

func WithFileAttachment(file ...string) Builder {
	return func(b *message) {
		b.fileAttachments = file
	}
}

func WithFileInlines(line ...string) Builder {
	return func(b *message) {
		b.fileInlines = line
	}
}

type message struct {
	from    string
	to      []string
	subject string
	body    []byte
	html    []byte

	cc      []string
	bcc     []string
	replyTo []string

	headers map[string]string

	fileAttachments []string
	fileInlines     []string
}

func (b message) Encode(host string) ([]byte, error) {
	fromAddress, err := mail.ParseAddress(b.from)
	if err != nil {
		return nil, err
	}

	toAddresses, err := b.parserMailAddresses(b.to)
	if err != nil {
		return nil, err
	}

	msg := enmime.
		Builder().
		From(fromAddress.Name, fromAddress.Address).
		ToAddrs(toAddresses).
		Subject(b.subject).
		Text(b.body).
		Date(time.Now().UTC())

	// Add cc
	ccAddresses, err := b.parserMailAddresses(b.cc)
	if err != nil {
		return nil, err
	}

	msg = msg.CCAddrs(ccAddresses)

	// Add bcc
	bccAddresses, err := b.parserMailAddresses(b.bcc)
	if err != nil {
		return nil, err
	}

	msg = msg.BCCAddrs(bccAddresses)

	// Add replyTo
	replyToAddresses, err := b.parserMailAddresses(b.replyTo)
	if err != nil {
		return nil, err
	}

	msg = msg.ReplyToAddrs(replyToAddresses)

	// Add headers
	if b.headers == nil {
		b.headers = make(map[string]string)
	}

	if _, ok := b.headers["Message-Id"]; !ok {
		b.headers["Message-Id"] = fmt.Sprintf("<%s@%s>", goutils.GenRandomString(32), host)
	}

	for k, v := range b.headers {
		msg = msg.Header(k, v)
	}

	if len(b.html) > 0 {
		msg = msg.HTML(b.html)
	}

	// Add attachment files
	for _, file := range b.fileAttachments {
		msg = msg.AddFileAttachment(file)
	}

	// Add inlines
	for _, line := range b.fileInlines {
		msg = msg.AddFileInline(line)
	}

	part, err := msg.Build()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = part.Encode(&buf)

	return buf.Bytes(), err
}

func (b message) parserMailAddresses(lists []string) (addresses []mail.Address, err error) {
	for _, addr := range lists {
		address, err := mail.ParseAddress(addr)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, *address)
	}

	return
}
