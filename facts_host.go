package goutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type OSInfo struct {
	System       string `json:"system"` // linux, darwin, freebsd, openbsd, windows
	OSFamily     string `json:"os_family"`
	Architecture string `json:"architecture"` // 386, amd64, arm, arm64
	Arch         string `json:"arch"`

	// Distribution
	// - Debian, Ubuntu
	// - RedHat, CentOS, Rocky, AlmaLinux
	// - OpenBSD
	Distribution string `json:"distribution"`

	// DistributionRelease
	// - CentOS Stream -> Stream
	// - Rocky -> Rocky
	// - AlmaLinux -> AlmaLinux
	DistributionRelease      string `json:"distribution_release"`
	DistributionVersion      string `json:"distribution_version"`
	DistributionMajorVersion string `json:"distribution_major_version"`
}

func (oi OSInfo) ToMap() (m map[string]string, err error) {
	jb, err := json.Marshal(oi)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jb, &m)

	return
}

func GatherOSInfo() (oi OSInfo, err error) {
	oi.System = runtime.GOOS // linux, darwin, freebsd, openbsd, windows
	oi.OSFamily = runtime.GOOS
	oi.Architecture = runtime.GOARCH // 386, amd64, arm, arm64
	oi.Arch = runtime.GOARCH

	if oi.System == "linux" {
		// Check file `/etc/os-release` to get distribution and release
		// https://www.freedesktop.org/software/systemd/man/os-release.html
		fpth := "/etc/os-release"

		if _, err = os.Stat(fpth); os.IsNotExist(err) {
			err = fmt.Errorf("file %s does not exist on linux machine", fpth)

			return
		}

		content, err1 := os.ReadFile(fpth)
		if err1 != nil {
			err = fmt.Errorf("failed read content of file %s", fpth)

			return
		}

		m := make(map[string]string)
		for _, line := range strings.Split(string(content), "\n") {
			if strings.Contains(line, "=") {
				items := strings.Split(line, "=")

				// Remove prefix/suffix quotes from values
				items[1] = strings.TrimPrefix(items[1], "\"")
				items[1] = strings.TrimSuffix(items[1], "\"")

				m[strings.ToLower(items[0])] = items[1]
			}
		}

		if v, ok := m["id"]; ok {
			oi.Distribution = v

			switch v {
			case "centos", "rocky", "almalinux":
				oi.OSFamily = "RedHat"
			case "debian", "ubuntu":
				oi.OSFamily = "Debian"
			}
		}

		if v, ok := m["version_id"]; ok {
			oi.DistributionVersion = v                             // "20.04"
			oi.DistributionMajorVersion = strings.Split(v, ".")[0] // "20"
		}

		if v, ok := m["version_codename"]; ok {
			// mostly used in Debian family
			oi.DistributionRelease = v // "focal"
		}

		if v, ok := m["name"]; ok {
			if v == "CentOS Stream" {
				oi.DistributionRelease = "Stream"
			}
		}

		return
	} else if runtime.GOOS == "openbsd" {
		var stdout bytes.Buffer
		cmd := exec.Command("uname", "-r")
		cmd.Stdout = &stdout
		if err = cmd.Run(); err != nil {
			return
		}

		oi.DistributionVersion = strings.TrimSpace(stdout.String())

		return
	}

	return
}
