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

type OsInfo struct {
	System                   string `json:"system,omitempty"`       // linux, darwin, freebsd, openbsd, windows
	OsFamily                 string `json:"os_family,omitempty"`    // Note: on Linux, it will be overwritten.
	Architecture             string `json:"architecture,omitempty"` // 386, amd64, arm, arm64
	Arch                     string `json:"arch,omitempty"`         // alias of `ansible_architecture`
	CronFileOwner            string `json:"cron_file_owner,omitempty"`
	CronFileGroup            string `json:"cron_file_group,omitempty"`
	CronSpoolDir             string `json:"cron_spool_dir,omitempty"`
	Distribution             string `json:"distribution,omitempty"`
	DistributionRelease      string `json:"distribution_release,omitempty"`
	DistributionVersion      string `json:"distribution_version,omitempty"`
	DistributionMajorVersion string `json:"distribution_major_version,omitempty"`
	PkgMgr                   string `json:"pkg_mgr,omitempty"`
}

func (oi OsInfo) ToMap() (map[string]any, error) {
	marshal, err := json.Marshal(oi)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	err = json.Unmarshal(marshal, &m)

	return m, err
}

func GatherOSInfo() (oi OsInfo, err error) {
	// OS
	oi.System = runtime.GOOS         // linux, darwin, freebsd, openbsd, windows
	oi.OsFamily = runtime.GOOS       // Note: on Linux, it will be overwritten.
	oi.Architecture = runtime.GOARCH // 386, amd64, arm, arm64
	oi.Arch = runtime.GOARCH         // alias of `ansible_architecture`

	// cron
	oi.CronFileOwner = "root"
	oi.CronFileGroup = "root"

	if oi.System == "linux" {
		// Check file `/etc/os-release` to get distribution and release
		// https://www.freedesktop.org/software/systemd/man/os-release.html
		fpth := "/etc/os-release"

		if _, err = os.Stat(fpth); os.IsNotExist(err) {
			err = fmt.Errorf("file %s does not exist on linux machine", fpth)

			return
		}

		contentBytes, err1 := os.ReadFile(fpth)
		if err1 != nil {
			err = fmt.Errorf("failed read content of file %s", fpth)

			return
		}

		tmpm := make(map[string]string)
		for _, line := range strings.Split(string(contentBytes), "\n") {
			if strings.Contains(line, "=") {
				items := strings.Split(line, "=")

				// Remove prefix/suffix quotes from values
				items[1] = strings.TrimPrefix(items[1], "\"")
				items[1] = strings.TrimSuffix(items[1], "\"")

				tmpm[strings.ToLower(items[0])] = items[1]
			}
		}

		var name string
		if val, ok := tmpm["name"]; ok {
			name = val
		}
		if val, ok := tmpm["id"]; ok {
			oi.Distribution = val // "ubuntu"
			oi.OsFamily = val
		}
		if val, ok := tmpm["version_id"]; ok {
			oi.DistributionVersion = val                             // "20.04"
			oi.DistributionMajorVersion = strings.Split(val, ".")[0] // "20"
		}
		if val, ok := tmpm["version_codename"]; ok {
			oi.DistributionRelease = val // "focal"
		}

		// FYI: https://github.com/ansible/ansible/blob/devel/lib/ansible/module_utils/facts/system/distribution.py#L470
		//	'RedHat': ['RedHat', 'Fedora', 'CentOS', 'Scientific', 'SLC',
		//              'Ascendos', 'CloudLinux', 'PSBM', 'OracleLinux', 'OVS',
		//              'OEL', 'Amazon', 'Virtuozzo', 'XenServer', 'Alibaba',
		//              'EulerOS', 'openEuler', 'AlmaLinux', 'Rocky'],
		//	'Debian': ['Debian', 'Ubuntu', 'Raspbian', 'Neon', 'KDE neon',
		//                                'Linux Mint', 'SteamOS', 'Devuan', 'Kali', 'Cumulus Linux',
		//                                'Pop!_OS', 'Parrot', 'Pardus GNU/Linux'],
		//  'Suse': ['SuSE', 'SLES', 'SLED', 'openSUSE', 'openSUSE Tumbleweed',
		//                              'SLES_SAP', 'SUSE_LINUX', 'openSUSE Leap'],
		//  'Archlinux': ['Archlinux', 'Antergos', 'Manjaro'],
		//  'Mandrake': ['Mandrake', 'Mandriva'],
		//  'Solaris': ['Solaris', 'Nexenta', 'OmniOS', 'OpenIndiana', 'SmartOS'],
		//  'Slackware': ['Slackware'],
		//  'Altlinux': ['Altlinux'],
		//  'Alpine': ['Alpine'],
		//  'Darwin': ['MacOSX'],
		//  'FreeBSD': ['FreeBSD', 'TrueOS'],
		//  'ClearLinux': ['Clear Linux OS', 'Clear Linux Mix'],
		//  'DragonFly': ['DragonflyBSD', 'DragonFlyBSD', 'Gentoo/DragonflyBSD', 'Gentoo/DragonFlyBSD'],
		//  'NetBSD': ['NetBSD'], }
		switch oi.Distribution {
		case "debian":
			oi.CronSpoolDir = "/var/spool/cron/crontabs"
			oi.CronFileGroup = "crontab"
			oi.PkgMgr = "apt"
		case "ubuntu":
			oi.PkgMgr = "apt"
		case "redhat":
			oi.CronSpoolDir = "/var/spool/cron"
			oi.PkgMgr = "dnf"
		case "centos":
			oi.PkgMgr = "dnf"
			if strings.Contains(name, "CentOS Stream") {
				oi.DistributionRelease = name
			}
		case "rocky":
			oi.PkgMgr = "dnf"
		case "almalinux":
			oi.PkgMgr = "dnf"
		}

		return
	}

	if runtime.GOOS == "openbsd" {
		var stdout bytes.Buffer
		command := exec.Command("uname", "-r")
		command.Stdout = &stdout
		if err = command.Run(); err != nil {
			return
		}

		oi.DistributionVersion = strings.TrimSpace(stdout.String())
		oi.PkgMgr = "openbsd_pkg"
		oi.CronSpoolDir = "/var/cron/tabs"
		oi.CronFileGroup = "crontab"

		return
	}

	if runtime.GOOS == "freebsd" {
		oi.PkgMgr = "freebsd_pkg"
		oi.CronSpoolDir = "/var/cron/tabs"
		oi.CronFileGroup = "crontab"
	}

	return
}
