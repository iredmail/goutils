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

type ResultDKIM struct {
	Domain string
	RTT    time.Duration
	TTL    uint32
	DKIM   string
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
