package doveadmpw

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func VerifyPassword(hash, plain string) (matched bool, err error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := exec.Command("doveadm", "pw", "-t", hash, "-p", plain)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		return false, fmt.Errorf("stdout: %s, stderr: %s, err: %w", stdout.String(), stderr.String(), err)
	}

	// Sample doveadm-pw output:
	//
	// - matched / verified:
	//
	//	$ doveadm pw -t '{SSHA}Ix...' -p 'HHiJ...'
	//	{SSHA}Ix... (verified)
	//
	// - mismatch:
	//
	//	$ doveadm pw -t '{SSHA}Ix...' -p 'HHiJ...'
	//	Fatal: reverse password verification check failed: Password mismatch
	//
	out := strings.TrimSpace(stdout.String())
	if strings.HasSuffix(out, "(verified)") {
		matched = true

		return
	}

	return
}

func IsSupportedPasswordScheme(pwHash string) bool {
	if !(strings.HasPrefix(pwHash, "{") && strings.Contains(pwHash, "}")) {
		return false
	}

	// Extract scheme name from password hash: "{SSHA}xxxx" -> "SSHA"
	scheme := strings.Split(strings.Split(pwHash, "}")[0], "{")[1]
	scheme = strings.ToUpper(scheme)

	supportedSchemes := []string{
		"PLAIN",
		"CRYPT",
		"MD5",
		"PLAIN-MD5",
		"SHA",
		"SSHA",
		"SHA512",
		"SSHA512",
		"SHA512-CRYPT",
		"BCRYPT",
		"CRAM-MD5",
		"NTLM",
	}

	for _, supportedScheme := range supportedSchemes {
		if scheme == supportedScheme {
			return true
		}
	}

	return false
}
