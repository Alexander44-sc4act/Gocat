// Package profiling provides memory profiling and optimization tools.
//
// The profiling package helps monitor and optimize memory usage in GoCat.
// It provides real-time memory statistics, allocation tracking, and
// memory threshold alerts.
//
// # Profiler
//
// Create and start a memory profiler:
//
//	profiler := profiling.NewProfiler(100, time.Second)
//	profiler.Start()
//	defer profiler.Stop()
//
// Get current memory statistics:
//
//	stats := profiler.GetCurrentStats()
//	fmt.Printf("Heap: %s\n", profiling.FormatBytes(stats.HeapAlloc))
//
// Generate a memory report:
//
//	report := profiler.GetMemoryReport()
//	fmt.Println(report)
//
// # Allocation Tracking
//
// Track memory allocations by name:
//
//	profiler.TrackAllocation("buffer", 4096)
//	profiler.TrackAllocation("connection", 1024)
//
// Get top allocations:
//
//	topAllocs := profiler.GetTopAllocations(10)
//	for _, alloc := range topAllocs {
//	    fmt.Printf("%s: %d bytes\n", alloc.Name, alloc.TotalBytes)
//	}
//
// # Memory Watcher
//
// Set up memory threshold alerts:
//
//	watcher := profiling.NewMemoryWatcher(
//	    100*1024*1024, // 100MB threshold
//	    func(stats profiling.MemoryStats) {
//	        log.Printf("Memory threshold exceeded: %s", profiling.FormatBytes(stats.HeapAlloc))
//	    },
//	    time.Second,
//	)
//	watcher.Start()
//
// # Best Practices
//
//   - Start profiling early in development
//   - Monitor heap allocations during load testing
//   - Use allocation tracking to identify memory hotspots
//   - Set up memory watchers for production monitoring
package profiling
