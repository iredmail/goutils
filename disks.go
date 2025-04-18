package goutils

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v4/disk"
)

type DiskInfo struct {
	Mounted     string `json:"mounted"`
	Total       string `json:"total"`
	Used        string `json:"used"`
	Free        string `json:"free"`
	UsedPercent int    `json:"used_percent"`
}

func GetDiskInfo() (dis []DiskInfo, err error) {
	var partitionStats []disk.PartitionStat
	partitionStats, err = disk.PartitionsWithContext(context.Background(), false)
	if err != nil {
		return
	}

	for _, p := range partitionStats {
		if strings.HasPrefix(p.Mountpoint, "/snap/") {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		if usage.Total == 0 {
			continue
		}

		dis = append(dis, DiskInfo{
			Mounted:     p.Mountpoint,
			Total:       humanize.IBytes(usage.Total),
			Used:        humanize.IBytes(usage.Used),
			Free:        humanize.IBytes(usage.Free),
			UsedPercent: int(usage.UsedPercent),
		})
	}

	slices.SortFunc(dis, func(a, b DiskInfo) int {
		return cmp.Compare(a.Mounted, b.Mounted)
	})

	return
}
