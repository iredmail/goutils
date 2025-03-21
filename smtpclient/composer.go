package smtpclient

import (
	"bytes"
	"net/mail"
	"time"

	"github.com/jhillyerd/enmime/v2"

	"github.com/iredmail/goutils"
)

// Composer 用于编写邮件。
type Composer struct {
	from     mail.Address
	to       []mail.Address
	subject  string // mail subject
	bodyText []byte // mail body in plain text format
	bodyHTML []byte // mail body in html format

	// Optional
	cc              []mail.Address // header `Cc:`
	bcc             []mail.Address // header `Bcc:`
	replyTo         mail.Address   // header `Reply-To:`
	date            time.Time      // header `Date:`
	messageID       string         // header `Message-ID:`
	headers         map[string]string
	fileAttachments []string // Path to files
	byteAttachments []*ByteAttachment
}

type ByteAttachment struct {
	Name        string
	ContentType string
	Bytes       []byte
}

func NewComposer() *Composer {
	return &Composer{
		headers: make(map[string]string),
	}
}

func (c *Composer) WithFrom(from mail.Address) *Composer {
	c.from = from

	return c
}

func (c *Composer) WithTo(to []mail.Address) *Composer {
	c.to = to

	return c
}

func (c *Composer) WithSubject(subject string) *Composer {
	c.subject = subject

	return c
}

func (c *Composer) WithBodyText(text []byte) *Composer {
	c.bodyText = text

	return c
}

func (c *Composer) WithBodyHTML(html []byte) *Composer {
	c.bodyHTML = html

	return c
}

func (c *Composer) WithCc(cc []mail.Address) *Composer {
	c.cc = cc

	return c
}

func (c *Composer) WithBcc(bcc []mail.Address) *Composer {
	c.bcc = bcc

	return c
}

func (c *Composer) WithReplyTo(replyTo mail.Address) *Composer {
	c.replyTo = replyTo

	return c
}

func (c *Composer) WithDate(t time.Time) *Composer {
	c.date = t

	return c
}

func (c *Composer) WithMessageID(msgid string) *Composer {
	c.messageID = msgid

	return c
}

func (c *Composer) WithHeaders(headers map[string]string) *Composer {
	c.headers = headers

	return c
}

func (c *Composer) WithFileAttachments(pths ...string) *Composer {
	c.fileAttachments = pths

	return c
}

func (c *Composer) WithByteAttachments(atts ...*ByteAttachment) *Composer {
	c.byteAttachments = atts

	return c
}

// Bytes 将邮件内容转换为 `[]byte`。
func (c *Composer) Bytes() (msg []byte, err error) {
	mb := enmime.Builder().
		From(c.from.Name, c.from.Address).
		ToAddrs(c.to).
		Subject(c.subject)

	if len(c.bodyText) > 0 {
		mb = mb.Text(c.bodyText)
	}

	if len(c.bodyHTML) > 0 {
		mb = mb.HTML(c.bodyHTML)
	}

	if c.cc != nil {
		mb = mb.CCAddrs(c.cc)
	}

	if c.cc != nil {
		mb = mb.CCAddrs(c.cc)
	}

	if len(c.replyTo.Address) > 0 {
		mb = mb.ReplyTo(c.replyTo.Name, c.replyTo.Address)
	}

	if c.date.UTC().Unix() > 0 {
		mb = mb.Date(c.date)
	} else {
		mb = mb.Date(time.Now().UTC())
	}

	if len(c.messageID) > 0 {
		c.headers["Message-ID"] = c.messageID
	} else {
		// Use random Message-ID
		c.headers["Message-ID"] = goutils.GenRandomString(32) + "@" + goutils.GetHostFQDN()
	}

	if c.headers != nil {
		for k, v := range c.headers {
			mb = mb.Header(k, v)
		}
	}

	if c.fileAttachments != nil {
		for _, pth := range c.fileAttachments {
			mb = mb.AddFileAttachment(pth)
		}
	}

	if c.byteAttachments != nil {
		for _, att := range c.byteAttachments {
			mb = mb.AddAttachment(att.Bytes, att.ContentType, att.Name)
		}
	}

	part, err := mb.Build()
	if err != nil {
		return
	}

	var buf bytes.Buffer
	err = part.Encode(&buf)
	if err != nil {
		return
	}

	msg = buf.Bytes()

	return
}

func (c *Composer) GetTo() []mail.Address  { return c.to }
func (c *Composer) GetCc() []mail.Address  { return c.cc }
func (c *Composer) GetBcc() []mail.Address { return c.bcc }
func (c *Composer) GetSubject() string     { return c.subject }

func (c *Composer) GetAllRecipients() (addrs []mail.Address) {
	addrs = c.GetTo()
	addrs = append(addrs, c.GetCc()...)
	addrs = append(addrs, c.GetBcc()...)

	return
}
