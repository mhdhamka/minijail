package runtime

import (
	"fmt"
	"math/rand"
	"sync"
)

// CgroupController simulates and applies Linux cgroups v2 resource accounting
type CgroupController struct {
	mu sync.RWMutex
}

// NewCgroupController creates a new cgroups controller
func NewCgroupController() *CgroupController {
	return &CgroupController{}
}

// InitLimits constructs initial CgroupLimits struct
func (c *CgroupController) InitLimits(id string, memoryMB int64, cpuCores float64) CgroupLimits {
	if memoryMB <= 0 {
		memoryMB = 128
	}
	if cpuCores <= 0 {
		cpuCores = 1.0
	}

	memBytes := memoryMB * 1024 * 1024
	periodUs := int64(100000) // standard 100ms CFS period
	quotaUs := int64(cpuCores * float64(periodUs))

	// Initial baseline memory usage (e.g. 8-16 MB for minimal runtime/sh)
	baselineMem := int64(8+rand.Intn(10)) * 1024 * 1024

	return CgroupLimits{
		CgroupPath:       fmt.Sprintf("/sys/fs/cgroup/minijail/%s", id),
		MemoryMaxBytes:   memBytes,
		MemoryUsageBytes: baselineMem,
		MemoryMaxUsage:   baselineMem,
		CpuQuotaUs:       quotaUs,
		CpuPeriodUs:      periodUs,
		CpuShares:        1024,
		CpuPercent:       float64(rand.Intn(15)) + 5.0,
		ThrottlePeriods:  0,
		ThrottledTimeUs:  0,
		OomKillEvents:    0,
		OomScoreAdj:      0,
	}
}

// UpdateUsage simulates live cgroups hardware metrics tick
func (c *CgroupController) UpdateUsage(limits *CgroupLimits, isRunning bool, isStressed bool) {
	if !isRunning {
		limits.CpuPercent = 0
		limits.MemoryUsageBytes = 0
		return
	}

	// Calculate CPU usage within CFS quota
	maxAllowedPercent := (float64(limits.CpuQuotaUs) / float64(limits.CpuPeriodUs)) * 100.0

	if isStressed {
		// CPU is pegged up to quota limit, triggering CFS throttle
		limits.CpuPercent = maxAllowedPercent
		limits.ThrottlePeriods += int64(rand.Intn(5) + 3)
		limits.ThrottledTimeUs += int64(rand.Intn(12000) + 5000)
	} else {
		// Fluctuating normal process usage
		fluctuation := float64(rand.Intn(10)-5) / 2.0
		newPercent := limits.CpuPercent + fluctuation
		if newPercent < 2.0 {
			newPercent = 2.5
		}
		if newPercent > maxAllowedPercent {
			newPercent = maxAllowedPercent
			limits.ThrottlePeriods += 1
			limits.ThrottledTimeUs += 1500
		}
		limits.CpuPercent = newPercent
	}

	// Update peak memory
	if limits.MemoryUsageBytes > limits.MemoryMaxUsage {
		limits.MemoryMaxUsage = limits.MemoryUsageBytes
	}
}

// AddMemory simulates an allocation in the container's memory cgroup
// Returns true if OOM-kill was triggered!
func (c *CgroupController) AddMemory(limits *CgroupLimits, deltaMB int64) bool {
	deltaBytes := deltaMB * 1024 * 1024
	limits.MemoryUsageBytes += deltaBytes

	if limits.MemoryUsageBytes > limits.MemoryMaxUsage {
		limits.MemoryMaxUsage = limits.MemoryUsageBytes
	}

	// Check if we breached memory.max
	if limits.MemoryUsageBytes >= limits.MemoryMaxBytes {
		limits.OomKillEvents++
		limits.MemoryUsageBytes = limits.MemoryMaxBytes
		return true
	}
	return false
}
