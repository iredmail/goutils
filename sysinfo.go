package goutils

import (
	"fmt"
	"slices"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
)

type HostInfo struct {
	Hostname      string
	Architecture  string
	Memory        uint64
	OsName        string
	OsVersion     string
	BootTime      uint64
	KernelVersion string
	IPs           []string
	MACs          []string
	CPUList       []string

	// Uptime
	UptimeDays    uint64
	UptimeHours   uint64
	UptimeMinutes uint64
}

func GetAllAddresses() (IPs []string, MACs []string) {
	sts, err := psnet.Interfaces()
	if err != nil {
		return nil, nil
	}

	for _, st := range sts {
		for _, a := range st.Addrs {
			if len(a.Addr) > 0 {
				IPs = append(IPs, a.Addr)
			}

			if len(st.HardwareAddr) > 0 {
				MACs = append(MACs, st.HardwareAddr)
			}
		}
	}

	slices.Sort(IPs)
	slices.Sort(MACs)

	return
}

func GetSysHostInfo() (hi HostInfo, err error) {
	hi = HostInfo{}

	info, err := host.Info()
	if err != nil {
		return hi, nil
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return hi, nil
	}

	for _, ci := range cpuInfo {
		hi.CPUList = append(hi.CPUList, fmt.Sprintf("%s %.2fGHz (%d cores)", ci.ModelName, ci.Mhz/1000, ci.Cores))
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return hi, nil
	}

	uptimeSeconds := info.Uptime
	upDays := uptimeSeconds / (24 * 60 * 60)
	upHours := (uptimeSeconds % (24 * 60 * 60)) / (60 * 60)
	upMinutes := (uptimeSeconds % (60 * 60)) / 60
	ips, macs := GetAllAddresses()

	hi.Hostname = GetHostFQDN()
	hi.Architecture = info.KernelArch
	hi.Memory = vm.Total
	hi.OsName = info.Platform
	hi.OsVersion = info.PlatformVersion
	hi.BootTime = info.BootTime
	hi.IPs = ips
	hi.KernelVersion = info.KernelVersion
	hi.MACs = macs
	hi.UptimeDays = upDays
	hi.UptimeHours = upHours
	hi.UptimeMinutes = upMinutes

	return hi, nil
}
