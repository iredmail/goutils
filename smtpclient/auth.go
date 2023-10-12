package smtpclient

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"
)

func newLoginAuth(host, username, password string) *login {
	return &login{
		host:     host,
		username: username,
		password: password,
	}
}

type login struct {
	host     string
	username string
	password string
}

func (l *login) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		advertised := false
		for _, mechanism := range server.Auth {
			if mechanism == "LOGIN" {
				advertised = true
				break
			}
		}

		if !advertised {
			return "", nil, errors.New("unencrypted connection")
		}
	}

	if server.Name != l.host {
		return "", nil, errors.New("wrong hostname")
	}

	return "LOGIN", nil, nil
}

func (l *login) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	switch {
	case bytes.Equal(fromServer, []byte("Username:")):
		return []byte(l.username), nil
	case bytes.Equal(fromServer, []byte("Password:")):
		return []byte(l.password), nil
	default:
		return nil, fmt.Errorf("unexpected server challenge: %s", fromServer)
	}
}
