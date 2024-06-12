package goutils

import (
	"net"
	"strings"
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

func IsWildcardAddr(addr string) bool {
	ip := net.ParseIP(addr)
	if ip == nil {
		return false // Invalid IP address format
	}

	return ip.IsUnspecified()
}

func IsWildcardIPV4(addr string) bool {
	ip := net.ParseIP(addr)
	if ip == nil {
		return false // Invalid IP address format
	}

	if ip.To4() == nil {
		return false
	}

	return ip.IsUnspecified()
}

// GetIPPortFromNetAddrString 从格式为 `ip:port` （常用的是 `net.Addr.String()`）的字符串里获取 IP 和端口号。
func GetIPPortFromNetAddrString(addr string) (ip string, port string, version int) {
	slice := strings.Split(addr, ":")

	// IPv6 地址含有多个冒号，不能使用 slice[0], slice[1] 来获取 ip、port。
	port = slice[len(slice)-1]
	ip = strings.TrimSuffix(addr, ":"+port)

	version = 4
	if strings.Contains(ip, ":") {
		// IPv6. Strip `[]`.
		ip = strings.TrimPrefix(ip, "[")
		ip = strings.TrimSuffix(ip, "]")
		version = 6
	}

	return
}

// GetIPPortFromNetAddr 从 `net.Addr` 获取 IP 和端口号。
func GetIPPortFromNetAddr(addr net.Addr) (ip string, port string, version int) {
	ip, port, version = GetIPPortFromNetAddrString(addr.String())

	return
}
