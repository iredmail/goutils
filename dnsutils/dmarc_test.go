package dnsutils

import (
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/assert"
)

func TestQueryDMARC(t *testing.T) {
	domain := "gmail.com"
	found, result, err := QueryDMARC(domain)
	assert.True(t, found)
	assert.Nil(t, err)
	assert.Equal(t, domain, result.Domain)
	assert.Contains(t, result.Txt, "v=DMARC1")
	assert.Equal(t, "none", result.Params["p"])
	assert.Equal(t, "quarantine", result.Params["sp"])

	domain = "iredmail.org"
	found, result, err = QueryDMARC(domain)
	pp.Println(found, result, err)
	assert.True(t, true)
	assert.Nil(t, err)
	assert.Equal(t, domain, result.Domain)
	assert.Contains(t, result.Txt, "v=DMARC1")
	assert.Equal(t, "reject", result.Params["p"])
	assert.Equal(t, "s", result.Params["adkim"])
	assert.Equal(t, "s", result.Params["aspf"])
	assert.Equal(t, "100", result.Params["pct"])
	assert.Equal(t, "1", result.Params["fo"])
}
