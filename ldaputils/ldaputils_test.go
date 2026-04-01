package ldaputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractLDAPSuffixFromDN(t *testing.T) {
	// standard base dn with cn prefix
	assert.Equal(t, "dc=example,dc=com", ExtractLDAPSuffixFromDN("cn=abc,dc=example,dc=com"))
	// uppercase input should be converted to lowercase
	assert.Equal(t, "dc=example,dc=com", ExtractLDAPSuffixFromDN("CN=ABC,DC=EXAMPLE,DC=COM"))
	// mixed case input
	assert.Equal(t, "dc=test,dc=org", ExtractLDAPSuffixFromDN("Cn=test,Dc=Test,Dc=Org"))
	// input with leading and trailing whitespace
	assert.Equal(t, "dc=domain,dc=local", ExtractLDAPSuffixFromDN("  cn=user,dc=domain,dc=local  "))
	// input with whitespace around components (no match due to space before dc)
	assert.Equal(t, "cn=user , dc=domain , dc=local", ExtractLDAPSuffixFromDN("cn=user , dc=domain , dc=local"))
	// base dn with only dc components
	assert.Equal(t, "dc=example,dc=com", ExtractLDAPSuffixFromDN("dc=example,dc=com"))
	// base dn with single dc component
	assert.Equal(t, "dc=com", ExtractLDAPSuffixFromDN("dc=com"))
	// complex base dn with multiple ou and cn
	assert.Equal(t, "dc=company,dc=example,dc=com", ExtractLDAPSuffixFromDN("cn=admin,ou=users,ou=departments,dc=company,dc=example,dc=com"))
	// base dn with numeric values
	assert.Equal(t, "dc=test123,dc=org456", ExtractLDAPSuffixFromDN("cn=user123,dc=test123,dc=org456"))
	// base dn with hyphen and underscore
	assert.Equal(t, "dc=example-test,dc=com", ExtractLDAPSuffixFromDN("cn=user-name_test,dc=example-test,dc=com"))
	// empty string
	assert.Equal(t, "", ExtractLDAPSuffixFromDN(""))
	// whitespace only
	assert.Equal(t, "", ExtractLDAPSuffixFromDN("   "))
	// dn without dc component
	assert.Equal(t, "cn=user,ou=users", ExtractLDAPSuffixFromDN("cn=user,ou=users"))
	// dn with dc in cn value
	assert.Equal(t, "dc=example,dc=com", ExtractLDAPSuffixFromDN("cn=dc=test,dc=example,dc=com"))
}
