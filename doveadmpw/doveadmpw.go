package doveadmpw

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func GeneratePassword(scheme, password string) (hash string, err error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := exec.Command("doveadm", "pw", "-s", scheme, "-p", password)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("stdout: %s, stderr: %s, err: %w", stdout.String(), stderr.String(), err)

		return
	}

	// output: {SSHA}DVRj4taRESdmMKQ5oaCs69t7D3ZkHtMk
	hash = stdout.String()
	hash = strings.TrimRight(hash, "\n")

	return
}

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
