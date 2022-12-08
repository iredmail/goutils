package goutils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type osRelease struct {
	ID              string `json:"id"` // ID. OS name in lower case. "ubuntu"
	Name            string `json:"name"`
	VersionID       string `json:"version_id"`       // VERSION_ID. "20.04"
	VersionCodename string `json:"version_codename"` // VERSION_CODENAME. "focal"
}

func GatherOSInfo() (m map[string]any, err error) {
	m = make(map[string]any)

	// OS
	m["system"] = runtime.GOOS         // linux, darwin, freebsd, openbsd, windows
	m["os_family"] = runtime.GOOS      // Note: on Linux, it will be overwritten.
	m["architecture"] = runtime.GOARCH // 386, amd64, arm, arm64
	m["arch"] = runtime.GOARCH         // alias of `ansible_architecture`

	// cron
	m["cron_file_owner"] = "root"
	m["cron_file_group"] = "root"

	if runtime.GOOS == "linux" {
		// Check file `/etc/os-release` to get distribution and release
		// https://www.freedesktop.org/software/systemd/man/os-release.html
		fpth := "/etc/os-release"

		if _, err = os.Stat(fpth); os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s does not exist on linux machine", fpth)
		}

		contentBytes, err := os.ReadFile(fpth)
		if err != nil {
			return nil, fmt.Errorf("failed read content of file %s", fpth)
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

		var osm osRelease
		if val, ok := tmpm["name"]; ok {
			osm.Name = val
		}
		if val, ok := tmpm["id"]; ok {
			osm.ID = val
		}
		if val, ok := tmpm["version_id"]; ok {
			osm.VersionID = val
		}
		if val, ok := tmpm["version_codename"]; ok {
			osm.VersionCodename = val
		}

		m["distribution"] = osm.ID // "ubuntu"
		m["os_family"] = osm.ID
		m["distribution_release"] = osm.VersionCodename                        // "focal"
		m["distribution_version"] = osm.VersionID                              // "20.04"
		m["distribution_major_version"] = strings.Split(osm.VersionID, ".")[0] // "20"

		// Default value, will be overwritten later.
		m["cron_spool_dir"] = ""

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
		switch osm.ID {
		case "debian":
			m["pkg_mgr"] = "apt"
		case "ubuntu":
			m["pkg_mgr"] = "apt"
		case "redhat":
			m["pkg_mgr"] = "dnf"
		case "centos":
			m["pkg_mgr"] = "dnf"

			if strings.Contains(osm.Name, "CentOS Stream") {
				m["distribution_release"] = osm.Name
			}
		case "rocky":
			m["pkg_mgr"] = "dnf"
		case "almalinux":
			m["pkg_mgr"] = "dnf"
		}

		switch m["os_family"] {
		case "redhat":
			m["cron_spool_dir"] = "/var/spool/cron"
		case "debian":
			m["cron_spool_dir"] = "/var/spool/cron/crontabs"
			m["cron_file_group"] = "crontab"
		}
	} else if runtime.GOOS == "openbsd" {
		var stdout bytes.Buffer
		command := exec.Command("uname", "-r")
		command.Stdout = &stdout
		if err = command.Run(); err != nil {
			return nil, err
		}

		m["distribution_version"] = strings.TrimSpace(stdout.String())
		m["pkg_mgr"] = "openbsd_pkg"
		m["cron_spool_dir"] = "/var/cron/tabs"
		m["cron_file_group"] = "crontab"
	} else if runtime.GOOS == "freebsd" {
		m["pkg_mgr"] = "freebsd_pkg"
		m["cron_spool_dir"] = "/var/cron/tabs"
		m["cron_file_group"] = "crontab"
	}

	return
}
