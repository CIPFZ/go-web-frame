package api

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

type StateApi struct {
	svcCtx *svc.ServiceContext
}

func NewStateApi(svcCtx *svc.ServiceContext) *StateApi {
	return &StateApi{svcCtx: svcCtx}
}

type osInfo struct {
	Goos         string `json:"goos"`
	NumCpu       int    `json:"numCpu"`
	Compiler     string `json:"compiler"`
	GoVersion    string `json:"goVersion"`
	NumGoroutine int    `json:"numGoroutine"`
}

type cpuInfo struct {
	Cpus   []float64 `json:"cpus"`
	Cores  int       `json:"cores"`
	Load1  float64   `json:"load1"`
	Load5  float64   `json:"load5"`
	Load15 float64   `json:"load15"`
}

type ramInfo struct {
	Used        uint64  `json:"used"`
	Total       uint64  `json:"total"`
	UsedPercent float64 `json:"usedPercent"`
}

type diskInfo struct {
	MountPoint  string  `json:"mountPoint"`
	Used        uint64  `json:"used"`
	Total       uint64  `json:"total"`
	UsedPercent float64 `json:"usedPercent"`
	ReadBytes   uint64  `json:"readBytes"`
	WriteBytes  uint64  `json:"writeBytes"`
}

type ioInfo struct {
	ReadBytes  uint64 `json:"readBytes"`
	WriteBytes uint64 `json:"writeBytes"`
}

type serverInfo struct {
	OS   osInfo     `json:"os"`
	CPU  cpuInfo    `json:"cpu"`
	RAM  ramInfo    `json:"ram"`
	Disk []diskInfo `json:"disk"`
	IO   ioInfo     `json:"io"`
}

// GetServerInfo returns runtime, cpu, memory and disk metrics for backend host/container.
func (a *StateApi) GetServerInfo(c *gin.Context) {
	cpuPercents, err := cpu.Percent(0, true)
	if err != nil {
		response.FailWithMessage("获取 CPU 信息失败: "+err.Error(), c)
		return
	}
	physicalCores, err := cpu.Counts(false)
	if err != nil {
		physicalCores = 0
	}

	loadAvg, _ := load.Avg()

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		response.FailWithMessage("获取内存信息失败: "+err.Error(), c)
		return
	}

	partitions, err := disk.Partitions(true)
	if err != nil {
		response.FailWithMessage("获取磁盘分区失败: "+err.Error(), c)
		return
	}

	ioCounters, _ := disk.IOCounters()
	var totalReadBytes uint64
	var totalWriteBytes uint64
	for _, io := range ioCounters {
		totalReadBytes += io.ReadBytes
		totalWriteBytes += io.WriteBytes
	}

	skipFSTypes := map[string]struct{}{
		"proc": {}, "sysfs": {}, "tmpfs": {}, "devtmpfs": {}, "devpts": {},
		"mqueue": {}, "cgroup": {}, "cgroup2": {}, "squashfs": {}, "nsfs": {},
	}
	skipMountPoints := map[string]struct{}{
		"/etc/hostname":    {},
		"/etc/hosts":       {},
		"/etc/resolv.conf": {},
	}

	diskInfos := make([]diskInfo, 0, len(partitions))
	seen := map[string]struct{}{}
	for _, p := range partitions {
		if _, ok := seen[p.Mountpoint]; ok {
			continue
		}
		seen[p.Mountpoint] = struct{}{}

		if _, ok := skipMountPoints[p.Mountpoint]; ok {
			continue
		}
		if _, ok := skipFSTypes[p.Fstype]; ok && p.Mountpoint != "/" {
			continue
		}
		if strings.HasPrefix(p.Mountpoint, "/proc") || strings.HasPrefix(p.Mountpoint, "/sys") || strings.HasPrefix(p.Mountpoint, "/dev") || strings.HasPrefix(p.Mountpoint, "/run") {
			continue
		}

		stat, statErr := os.Stat(p.Mountpoint)
		if p.Mountpoint != "/" && (statErr != nil || !stat.IsDir()) {
			continue
		}

		usage, usageErr := disk.Usage(p.Mountpoint)
		if usageErr != nil || usage.Total == 0 {
			continue
		}

		var readBytes uint64
		var writeBytes uint64
		if io, ok := ioCounters[p.Device]; ok {
			readBytes = io.ReadBytes
			writeBytes = io.WriteBytes
		} else if devBase := filepath.Base(p.Device); devBase != "" {
			if io, ok := ioCounters[devBase]; ok {
				readBytes = io.ReadBytes
				writeBytes = io.WriteBytes
			}
		} else if io, ok := ioCounters[p.Mountpoint]; ok {
			readBytes = io.ReadBytes
			writeBytes = io.WriteBytes
		}

		diskInfos = append(diskInfos, diskInfo{
			MountPoint:  p.Mountpoint,
			Used:        usage.Used,
			Total:       usage.Total,
			UsedPercent: usage.UsedPercent,
			ReadBytes:   readBytes,
			WriteBytes:  writeBytes,
		})
	}

	sort.Slice(diskInfos, func(i, j int) bool {
		return diskInfos[i].MountPoint < diskInfos[j].MountPoint
	})

	if len(diskInfos) == 0 {
		if usage, usageErr := disk.Usage("/"); usageErr == nil && usage.Total > 0 {
			diskInfos = append(diskInfos, diskInfo{
				MountPoint:  "/",
				Used:        usage.Used,
				Total:       usage.Total,
				UsedPercent: usage.UsedPercent,
			})
		}
	}

	info := serverInfo{
		OS: osInfo{
			Goos:         runtime.GOOS,
			NumCpu:       runtime.NumCPU(),
			Compiler:     runtime.Compiler,
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
		},
		CPU: cpuInfo{
			Cpus:  cpuPercents,
			Cores: physicalCores,
			Load1: func() float64 {
				if loadAvg == nil {
					return 0
				}
				return loadAvg.Load1
			}(),
			Load5: func() float64 {
				if loadAvg == nil {
					return 0
				}
				return loadAvg.Load5
			}(),
			Load15: func() float64 {
				if loadAvg == nil {
					return 0
				}
				return loadAvg.Load15
			}(),
		},
		RAM: ramInfo{
			Used:        vmStat.Used,
			Total:       vmStat.Total,
			UsedPercent: vmStat.UsedPercent,
		},
		Disk: diskInfos,
		IO: ioInfo{
			ReadBytes:  totalReadBytes,
			WriteBytes: totalWriteBytes,
		},
	}

	response.OkWithData(gin.H{"server": info}, c)
}
