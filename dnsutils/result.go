package dnsutils

import (
	"sync"
	"time"
)

type ResultAll struct {
	mu sync.RWMutex

	Domain      string
	Duration    time.Duration
	ResultA     ResultA
	ResultAAAA  ResultAAAA
	ResultMX    ResultMX
	ResultSPF   ResultSPF
	ResultDKIM  ResultDKIM
	ResultDMARC ResultDMARC
}

// ResultA 定义 A 记录的查询结果。
type ResultA struct {
	Domain   string
	Duration time.Duration
	IPs      []string
	Error    error
}

type ResultAAAA struct {
	Domain   string
	Duration time.Duration
	IPs      []string
	Error    error
}

type HostMX struct {
	Hostname string   `json:"hostname"`
	IP4      []string `json:"ip4"`
	IP6      []string `json:"ip6"`
	Duration time.Duration
	TTL      uint32 `json:"ttl"`
	Priority uint16 `json:"priority"`
	Error    error
}

// ResultMX 定义 MX 记录的查询结果。
type ResultMX struct {
	Domain    string `json:"domain"`
	Duration  time.Duration
	Hostnames []string `json:"hostnames"`
	Hosts     []HostMX `json:"hosts"`
	Error     error
}

type ResultDKIM struct {
	Domain   string
	Duration time.Duration
	TTL      uint32
	DKIM     string
	Error    error
}

type ResultMTASTS struct {
	Domain   string
	Duration time.Duration
	TTL      uint32
	Error    error
}

type ResultTLSRPT struct {
	Domain   string
	Duration time.Duration
	TTL      uint32
	Error    error
}
