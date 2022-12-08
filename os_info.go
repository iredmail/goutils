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

	// Package manager
	PkgMgr string `json:"pkg_mgr"`
}

func (oi OSInfo) ToMap() (m map[string]string, err error) {
	jb, err := json.Marshal(oi)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jb, &m)

	return
}

func GetOSInfo() (oi OSInfo, err error) {
	oi.Architecture = runtime.GOARCH // 386, amd64, arm, arm64
	oi.Arch = runtime.GOARCH

	if runtime.GOOS == "linux" {
		oi.System = "Linux"

		// Check file `/etc/os-release` to get distribution and release
		// https://www.freedesktop.org/software/systemd/man/os-release.html
		fpth := "/etc/os-release"

		if _, err = os.Stat(fpth); os.IsNotExist(err) {
			err = fmt.Errorf("file %s does not exist on linux machine", fpth)

			return
		}

		var content []byte
		content, err = os.ReadFile(fpth)
		if err != nil {
			err = fmt.Errorf("failed in reading file: %s, %v", fpth, err)

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

		if id, ok := m["id"]; ok {
			switch id {
			case "debian":
				oi.Distribution = "Debian"
				oi.OSFamily = "Debian"
				oi.PkgMgr = "apt"
			case "ubuntu":
				oi.Distribution = "Ubuntu"
				oi.OSFamily = "Debian"
				oi.PkgMgr = "apt"
			case "redhat":
				oi.Distribution = "RedHat"
				oi.OSFamily = "RedHat"
				oi.PkgMgr = "dnf"
			case "centos":
				oi.Distribution = "CentOS"
				oi.OSFamily = "RedHat"
				oi.PkgMgr = "dnf"
			case "rocky":
				oi.Distribution = "Rocky"
				oi.OSFamily = "RedHat"
				oi.DistributionRelease = "Rocky"
				oi.PkgMgr = "dnf"
			case "almalinux":
				oi.Distribution = "AlmaLinux"
				oi.OSFamily = "RedHat"
				oi.DistributionRelease = "AlmaLinux"
				oi.PkgMgr = "dnf"
			}
		}

		if v, ok := m["version_id"]; ok {
			oi.DistributionVersion = v                             // "20.04"
			oi.DistributionMajorVersion = strings.Split(v, ".")[0] // "20"
		}

		if oi.OSFamily == "Debian" {
			if v, ok := m["version_codename"]; ok {
				oi.DistributionRelease = v // "focal"
			}
		}

		return
	} else if runtime.GOOS == "openbsd" {
		oi.System = "OpenBSD"
		oi.OSFamily = "OpenBSD"
		oi.Distribution = "OpenBSD"
		oi.PkgMgr = "openbsd_pkg"

		var stdout bytes.Buffer
		cmd := exec.Command("uname", "-r")
		cmd.Stdout = &stdout
		if err = cmd.Run(); err != nil {
			err = fmt.Errorf("error getting system info with command 'uname -r': %v", err)

			return
		}

		oi.DistributionVersion = strings.TrimSpace(stdout.String())

		return
	}

	return
}
