package goutils

import (
	"net"
)

func IsIP(s string) bool {
	return net.ParseIP(s) != nil
}

func IsCIDR(s string) bool {
	_, _, err := net.ParseCIDR(s)
	if err == nil {
		return true
	}

	return false
}
