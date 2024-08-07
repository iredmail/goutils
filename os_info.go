package goutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type OSInfo struct {
	Hostname string `json:"hostname"`
	HostID   string `json:"host_id"`

	System       string   `json:"system"`       // linux, darwin, freebsd, openbsd, windows
	Architecture string   `json:"architecture"` // 386, amd64, arm, arm64
	CPUCores     []string `json:"cpu_cores"`
	Memory       uint64   `json:"memory"` // in bytes
	Swap         uint64   `json:"swap"`   // in bytes

	// OS
	KernelVersion string `json:"kernel_version"`
	OSFamily      string `json:"os_family"`
	OSName        string `json:"os_name"`
	OSVersion     string `json:"os_version"`

	// Host is a docker container.
	IsContainer bool `json:"is_container"`

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

	// Uptime
	BootTime      uint64 `json:"boot_time"`
	UptimeDays    uint64 `json:"uptime_days"`
	UptimeHours   uint64 `json:"uptime_hours"`
	UptimeMinutes uint64 `json:"uptime_minutes"`

	// Net
	IPAddresses  []string `json:"ip_addresses"`
	MacAddresses []string `json:"mac_addresses"`

	// Metadata
	HasDovecotPgsqlLastLogin bool `json:"-"`
}

func (oi OSInfo) ToMap() (m map[string]any, err error) {
	jb, err := json.Marshal(oi)
	if err != nil {
		return nil, err
	}

	m = make(map[string]any)
	err = json.Unmarshal(jb, &m)

	return
}

// setDovecotPgsqlLastLogin 检查当前操作系统版本提供的 Dovecot 包是否支持将 last login time 保存在
// PostgreSQL 数据库。
//
// 注意：
//   - Dovecot 2.3.16 及后续版本才支持使用 PostgreSQL 存储 last login time。
//   - Dovecot 所有版本都支持使用 MySQL / MariaDB 存储 last login time。
//   - 此函数不需要 root 权限。
func (oi *OSInfo) setDovecotPgsqlLastLogin() {
	// 排除不支持的版本，后续的新版本都支持。
	switch oi.Distribution {
	case "Debian":
		// Debian 12 (Bookworm) 及后续版本都支持。
		if slices.Contains([]string{"10", "11"}, oi.DistributionVersion) {
			oi.HasDovecotPgsqlLastLogin = false
		}
	case "Ubuntu":
		// Ubuntu 22.04 及后续版本都支持。
		if slices.Contains([]string{"18.04", "20.04"}, oi.DistributionVersion) {
			oi.HasDovecotPgsqlLastLogin = false
		}
	case "RedHat", "CentOS", "Rocky", "AlmaLinux":
		// RHEL 9 及后续版本都支持。
		if slices.Contains([]string{"7", "8"}, oi.DistributionMajorVersion) {
			oi.HasDovecotPgsqlLastLogin = false
		}
	case "OpenBSD":
		// OpenBSD 7.3 及后续版本都支持。
		if slices.Contains([]string{"7.1", "7.2"}, oi.DistributionVersion) {
			oi.HasDovecotPgsqlLastLogin = false
		}
	}

	oi.HasDovecotPgsqlLastLogin = true
}

func GetOSInfo() (oi OSInfo, err error) {
	oi.Architecture = runtime.GOARCH // 386, amd64, arm, arm64

	//
	// CPU
	//
	cpuInfo, err := cpu.Info()
	if err != nil {
		return
	}

	for _, ci := range cpuInfo {
		oi.CPUCores = append(oi.CPUCores, fmt.Sprintf("%s %.2fGHz (%d cores)", ci.ModelName, ci.Mhz/1000, ci.Cores))
	}

	//
	// Memory
	//
	vm, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	oi.Memory = vm.Total

	sms, err := mem.SwapMemory()
	if err != nil {
		return
	}
	oi.Swap = sms.Total

	hi, err := host.Info()
	if err != nil {
		return
	}

	//
	// Hostname
	//
	oi.Hostname = hi.Hostname
	oi.HostID = hi.HostID
	oi.KernelVersion = hi.KernelVersion

	//
	// OS
	//
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
	}

	oi.OSName = hi.Platform
	oi.OSVersion = hi.PlatformVersion
	oi.setDovecotPgsqlLastLogin()

	// Docker container.
	if DestExists("/.dockerenv") {
		oi.IsContainer = true
	}

	// Network
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				return oi, err
			}
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					ip := ipNet.IP.String()
					if !slices.Contains(oi.IPAddresses, ip) {
						oi.IPAddresses = append(oi.IPAddresses, ip)
					}

					maddr := iface.HardwareAddr.String()
					if !slices.Contains(oi.MacAddresses, maddr) {
						oi.MacAddresses = append(oi.MacAddresses, maddr)
					}

					break
				}
			}
		}
	}

	uptime, err := host.Uptime()
	if err != nil {
		return
	}
	oi.UptimeDays = uptime / (60 * 60 * 24)
	oi.UptimeHours = (uptime - (oi.UptimeDays * 60 * 60 * 24)) / (60 * 60)
	oi.UptimeMinutes = ((uptime - (oi.UptimeDays * 60 * 60 * 24)) - (oi.UptimeHours * 60 * 60)) / 60

	return
}

// GetHostFQDN returns the FQDN returned by shell command `hostname -f` (linux/darwin) or `hostname` (openbsd).
func GetHostFQDN() (fqdn string) {
	var arg string
	var stdout bytes.Buffer

	switch runtime.GOOS {
	case "linux", "darwin":
		arg = "-f"
	}

	cmd := exec.Command("hostname", arg)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err == nil {
		fqdn = strings.TrimSpace(stdout.String())
	}

	return fqdn
}
