package dnsutils

// ResultA 定义 A 记录的查询结果。
type ResultA struct {
	Domain string
	IPs    []string
}

type ResultAAAA struct {
	Domain string
	IPs    []string
}

// ResultMX 定义 MX 记录的查询结果。
type ResultMX struct {
	Domain    string
	TTL       uint32
	Priority  uint16
	Hostnames []string
}

// ResultSPF 定义 SPF 记录的查询结果。
type ResultSPF struct {
	Domain string
	TTL    uint32
}

type ResultDKIM struct {
	Domain string
	TTL    uint32
}

type ResultDMARC struct {
	Domain string
	TTL    uint32
}

type ResultMTASTS struct {
	Domain string
	TTL    uint32
}

type ResultTLSRPT struct {
	Domain string
	TTL    uint32
}
