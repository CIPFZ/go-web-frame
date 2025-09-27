package copyguard

import (
	"go.uber.org/zap"
	"time"
)

// Config holds all the configurable parameters for a Manager.
type Config struct {
	// BandwidthLimit specifies the maximum transfer rate in bytes per second.
	// A value of 0 means no limit.
	BandwidthLimit float64

	// IOPSLimit specifies the maximum number of I/O operations per second.
	// A value of 0 means no limit.
	IOPSLimit float64

	// Concurrency specifies the maximum number of files to copy simultaneously.
	Concurrency int

	// CPUThreshold is the CPU usage percentage (0.0 to 1.0) above which
	// the copy manager will begin to dynamically reduce its speed.
	CPUThreshold float64

	// ReservedDiskSpace is the amount of disk space (in bytes) to keep free
	// on the destination volume. The manager will not write if the remaining
	// space falls below this value.
	ReservedDiskSpace uint64

	// MetricsSampleInterval is the frequency at which system and manager
	// metrics are collected.
	MetricsSampleInterval time.Duration

	// OnMetrics is an optional callback function that receives aggregated
	// metrics at regular intervals.
	OnMetrics func(metrics Metrics)

	// Logger is the zap logger instance for the manager to use.
	Logger *zap.Logger
}

// Metrics contains a snapshot of the manager's state and performance.
type Metrics struct {
	Timestamp time.Time
	// To be populated in later milestones
}

// Progress describes the real-time status of a single file copy operation.
type Progress struct {
	Src         string // Source file path
	Dst         string // Destination file path
	TotalBytes  int64  // Total size of the source file
	CopiedBytes int64  // Bytes copied so far
}
