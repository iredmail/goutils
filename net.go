package goutils

import (
	"net"
)

func IsIP(s string) bool {
	return net.ParseIP(s) != nil
}

func IsCIDR(s string) bool {
	_, _, err := net.ParseCIDR(s)

	return err == nil
}

func IsNetworkPort(num int) (ok bool) {
	if num > 0 && num <= 65535 {
		return true
	}

	return
}
