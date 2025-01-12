package dnsutils

import "time"

// ResultA 定义 A 记录的查询结果。
type ResultA struct {
	Domain string
	RTT    time.Duration
	IPs    []string
}

type ResultAAAA struct {
	Domain string
	RTT    time.Duration
	IPs    []string
}

type HostMX struct {
	Hostname string `json:"hostname"`
	RTT      time.Duration
	TTL      uint32 `json:"ttl"`
	Priority uint16 `json:"priority"`
}

// ResultMX 定义 MX 记录的查询结果。
type ResultMX struct {
	Domain    string `json:"domain"`
	RTT       time.Duration
	Hostnames []string `json:"hostnames"`
	Hosts     []HostMX `json:"hosts"`
}

// ResultSPF 定义 SPF 记录的查询结果。
type ResultSPF struct {
	Domain string
	RTT    time.Duration
	TTL    uint32
	Txt    string // 原始的 SPF 记录

	// IP4s     []string // `ip4:`
	// IP6s     []string // `ip6:`
	// As       []string // `a:`
	// MXs      []string // `mx:`
	// Includes []string // `include:`
	// 经过完整解析后得到的所有 IP 地址
	// AllIPs []string
}

type ResultDKIM struct {
	Domain string
	RTT    time.Duration
	TTL    uint32
	DKIM   string
}

type ResultDMARC struct {
	Domain string
	RTT    time.Duration
	TTL    uint32
}

type ResultMTASTS struct {
	Domain string
	RTT    time.Duration
	TTL    uint32
}

type ResultTLSRPT struct {
	Domain string
	RTT    time.Duration
	TTL    uint32
}
