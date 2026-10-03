// Package profiling provides memory profiling and optimization tools
package profiling

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryStats holds memory statistics
type MemoryStats struct {
	Alloc        uint64    `json:"alloc"`
	TotalAlloc   uint64    `json:"total_alloc"`
	Sys          uint64    `json:"sys"`
	NumGC        uint32    `json:"num_gc"`
	HeapAlloc    uint64    `json:"heap_alloc"`
	HeapSys      uint64    `json:"heap_sys"`
	HeapIdle     uint64    `json:"heap_idle"`
	HeapInuse    uint64    `json:"heap_inuse"`
	HeapReleased uint64    `json:"heap_released"`
	HeapObjects  uint64    `json:"heap_objects"`
	StackInuse   uint64    `json:"stack_inuse"`
	StackSys     uint64    `json:"stack_sys"`
	GCPauseNs    uint64    `json:"gc_pause_ns"`
	Goroutines   int       `json:"goroutines"`
	Timestamp    time.Time `json:"timestamp"`
}

// Profiler provides memory profiling capabilities
type Profiler struct {
	samples     []MemoryStats
	maxSamples  int
	interval    time.Duration
	running     atomic.Bool
	allocations map[string]*AllocationInfo
	mu          sync.RWMutex
	stopCh      chan struct{}
}

// AllocationInfo tracks allocation information
type AllocationInfo struct {
	Name       string
	Count      int64
	TotalBytes int64
	AvgBytes   int64
	LastAlloc  time.Time
}

// NewProfiler creates a new memory profiler
func NewProfiler(maxSamples int, interval time.Duration) *Profiler {
	if maxSamples <= 0 {
		maxSamples = 100
	}
	if interval <= 0 {
		interval = time.Second
	}
	return &Profiler{
		samples:     make([]MemoryStats, 0, maxSamples),
		maxSamples:  maxSamples,
		interval:    interval,
		allocations: make(map[string]*AllocationInfo),
		stopCh:      make(chan struct{}),
	}
}

// Start begins memory profiling
func (p *Profiler) Start() {
	if p.running.Swap(true) {
		return // Already running
	}

	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.collectSample()
			case <-p.stopCh:
				return
			}
		}
	}()
}

// Stop stops memory profiling
func (p *Profiler) Stop() {
	if p.running.Swap(false) {
		close(p.stopCh)
	}
}

// collectSample collects a memory sample
func (p *Profiler) collectSample() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := MemoryStats{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,
		StackInuse:   m.StackInuse,
		StackSys:     m.StackSys,
		Goroutines:   runtime.NumGoroutine(),
		Timestamp:    time.Now(),
	}

	if m.NumGC > 0 {
		stats.GCPauseNs = m.PauseNs[(m.NumGC+255)%256]
	}

	p.mu.Lock()
	p.samples = append(p.samples, stats)
	if len(p.samples) > p.maxSamples {
		p.samples = p.samples[1:]
	}
	p.mu.Unlock()
}

// GetCurrentStats returns current memory statistics
func (p *Profiler) GetCurrentStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,
		StackInuse:   m.StackInuse,
		StackSys:     m.StackSys,
		Goroutines:   runtime.NumGoroutine(),
		Timestamp:    time.Now(),
	}
}

// GetSamples returns collected memory samples
func (p *Profiler) GetSamples() []MemoryStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	samples := make([]MemoryStats, len(p.samples))
	copy(samples, p.samples)
	return samples
}

// TrackAllocation tracks a memory allocation
func (p *Profiler) TrackAllocation(name string, bytes int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if info, exists := p.allocations[name]; exists {
		info.Count++
		info.TotalBytes += bytes
		info.AvgBytes = info.TotalBytes / info.Count
		info.LastAlloc = time.Now()
	} else {
		p.allocations[name] = &AllocationInfo{
			Name:       name,
			Count:      1,
			TotalBytes: bytes,
			AvgBytes:   bytes,
			LastAlloc:  time.Now(),
		}
	}
}

// GetTopAllocations returns top N allocations by total bytes
func (p *Profiler) GetTopAllocations(n int) []*AllocationInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	allocs := make([]*AllocationInfo, 0, len(p.allocations))
	for _, info := range p.allocations {
		allocs = append(allocs, info)
	}

	sort.Slice(allocs, func(i, j int) bool {
		return allocs[i].TotalBytes > allocs[j].TotalBytes
	})

	if n > len(allocs) {
		n = len(allocs)
	}
	return allocs[:n]
}

// ForceGC forces garbage collection
func (p *Profiler) ForceGC() {
	runtime.GC()
}

// GetMemoryReport generates a memory report
func (p *Profiler) GetMemoryReport() string {
	stats := p.GetCurrentStats()
	return fmt.Sprintf(`Memory Report
=============
Heap Alloc:    %s
Heap Sys:      %s
Heap Objects:  %d
Stack Inuse:   %s
Goroutines:    %d
GC Cycles:     %d
Last GC Pause: %s
`,
		FormatBytes(stats.HeapAlloc),
		FormatBytes(stats.HeapSys),
		stats.HeapObjects,
		FormatBytes(stats.StackInuse),
		stats.Goroutines,
		stats.NumGC,
		time.Duration(stats.GCPauseNs),
	)
}

// FormatBytes formats bytes to human readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// MemoryWatcher watches for memory thresholds
type MemoryWatcher struct {
	threshold uint64
	callback  func(MemoryStats)
	interval  time.Duration
	running   atomic.Bool
	stopCh    chan struct{}
}

// NewMemoryWatcher creates a new memory watcher
func NewMemoryWatcher(threshold uint64, callback func(MemoryStats), interval time.Duration) *MemoryWatcher {
	return &MemoryWatcher{
		threshold: threshold,
		callback:  callback,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

// Start starts the memory watcher
func (w *MemoryWatcher) Start() {
	if w.running.Swap(true) {
		return
	}

	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				if m.HeapAlloc > w.threshold {
					stats := MemoryStats{
						HeapAlloc:  m.HeapAlloc,
						HeapSys:    m.HeapSys,
						Goroutines: runtime.NumGoroutine(),
						Timestamp:  time.Now(),
					}
					w.callback(stats)
				}
			case <-w.stopCh:
				return
			}
		}
	}()
}

// Stop stops the memory watcher
func (w *MemoryWatcher) Stop() {
	if w.running.Swap(false) {
		close(w.stopCh)
	}
}
