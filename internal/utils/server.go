package utils

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type Server struct {
	Os   Os     `json:"os"`
	Cpu  Cpu    `json:"cpu"`
	Ram  Ram    `json:"ram"`
	Disk []Disk `json:"disk"`
}

type Os struct {
	GOOS         string `json:"goos"`
	NumCPU       int    `json:"numCpu"`
	Compiler     string `json:"compiler"`
	GoVersion    string `json:"goVersion"`
	NumGoroutine int    `json:"numGoroutine"`
}

type Cpu struct {
	Cpus  []float64 `json:"cpus"`
	Cores int       `json:"cores"`
}

type Ram struct {
	Used  int `json:"used"`
	Total int `json:"total"`
}

type Disk struct {
	MountPoint string `json:"mountPoint"`
	Used       int    `json:"used"`
	Total      int    `json:"total"`
}

// InitOS OS信息
func InitOS() (o Os) {
	o.GOOS = runtime.GOOS
	o.NumCPU = runtime.NumCPU()
	o.Compiler = runtime.Compiler
	o.GoVersion = runtime.Version()
	o.NumGoroutine = runtime.NumGoroutine()
	return o
}

// InitCPU CPU信息
func InitCPU() (c Cpu, err error) {
	if cores, err := cpu.Counts(false); err != nil {
		return c, err
	} else {
		c.Cores = cores
	}
	if cpus, err := cpu.Percent(time.Duration(200)*time.Millisecond, true); err != nil {
		return c, err
	} else {
		c.Cpus = cpus
	}
	return c, nil
}

// InitRAM RAM信息
func InitRAM() (r Ram, err error) {
	if u, err := mem.VirtualMemory(); err != nil {
		return r, err
	} else {
		r.Used = int(u.Used)
		r.Total = int(u.Total)
	}
	return r, nil
}

// InitDisk 硬盘信息
func InitDisk(paths []string) (d []Disk, err error) {
	for i := range paths {
		if u, err := disk.Usage(paths[i]); err != nil {
			return d, err
		} else {
			d = append(d, Disk{
				MountPoint: paths[i],
				Used:       int(u.Used),
				Total:      int(u.Total),
			})
		}
	}
	return d, nil
}
