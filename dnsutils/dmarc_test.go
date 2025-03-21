package dnsutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryDMARC(t *testing.T) {
	domain := "gmail.com"
	found, result := QueryDMARC(domain)
	assert.True(t, found)
	assert.Nil(t, result.Error)
	assert.Equal(t, domain, result.Domain)
	assert.Contains(t, result.Txt, "v=DMARC1")
	assert.Equal(t, "none", result.Params["p"])
	assert.Equal(t, "quarantine", result.Params["sp"])

	domain = "iredmail.org"
	found, result = QueryDMARC(domain)
	assert.True(t, true)
	assert.Nil(t, result.Error)
	assert.Equal(t, domain, result.Domain)
	assert.Contains(t, result.Txt, "v=DMARC1")
	assert.Equal(t, "reject", result.Params["p"])
	assert.Equal(t, "s", result.Params["adkim"])
	assert.Equal(t, "s", result.Params["aspf"])
	assert.Equal(t, "100", result.Params["pct"])
	assert.Equal(t, "1", result.Params["fo"])
}
