package profiling

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewProfiler(t *testing.T) {
	p := NewProfiler(100, time.Second)

	if p == nil {
		t.Fatal("NewProfiler returned nil")
	}

	if p.maxSamples != 100 {
		t.Errorf("maxSamples = %d, want 100", p.maxSamples)
	}

	if p.interval != time.Second {
		t.Errorf("interval = %v, want 1s", p.interval)
	}
}

func TestNewProfilerDefaults(t *testing.T) {
	p := NewProfiler(0, 0)

	if p.maxSamples != 100 {
		t.Errorf("default maxSamples = %d, want 100", p.maxSamples)
	}

	if p.interval != time.Second {
		t.Errorf("default interval = %v, want 1s", p.interval)
	}
}

func TestGetCurrentStats(t *testing.T) {
	p := NewProfiler(10, time.Second)
	stats := p.GetCurrentStats()

	if stats.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	if stats.Goroutines <= 0 {
		t.Error("Goroutines should be > 0")
	}

	// HeapAlloc should be > 0 in any running Go program
	if stats.HeapAlloc == 0 {
		t.Error("HeapAlloc should be > 0")
	}
}

func TestTrackAllocation(t *testing.T) {
	p := NewProfiler(10, time.Second)

	p.TrackAllocation("test-buffer", 1024)
	p.TrackAllocation("test-buffer", 2048)
	p.TrackAllocation("other-buffer", 512)

	allocs := p.GetTopAllocations(10)

	if len(allocs) != 2 {
		t.Errorf("Expected 2 allocations, got %d", len(allocs))
	}

	// Find test-buffer
	var testBuffer *AllocationInfo
	for _, a := range allocs {
		if a.Name == "test-buffer" {
			testBuffer = a
			break
		}
	}

	if testBuffer == nil {
		t.Fatal("test-buffer not found")
	}

	if testBuffer.Count != 2 {
		t.Errorf("Count = %d, want 2", testBuffer.Count)
	}

	if testBuffer.TotalBytes != 3072 {
		t.Errorf("TotalBytes = %d, want 3072", testBuffer.TotalBytes)
	}

	if testBuffer.AvgBytes != 1536 {
		t.Errorf("AvgBytes = %d, want 1536", testBuffer.AvgBytes)
	}
}

func TestGetTopAllocations(t *testing.T) {
	p := NewProfiler(10, time.Second)

	p.TrackAllocation("small", 100)
	p.TrackAllocation("medium", 1000)
	p.TrackAllocation("large", 10000)

	allocs := p.GetTopAllocations(2)

	if len(allocs) != 2 {
		t.Errorf("Expected 2 allocations, got %d", len(allocs))
	}

	// Should be sorted by TotalBytes descending
	if allocs[0].Name != "large" {
		t.Errorf("First allocation should be 'large', got '%s'", allocs[0].Name)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, got, tt.want)
		}
	}
}

func TestGetMemoryReport(t *testing.T) {
	p := NewProfiler(10, time.Second)
	report := p.GetMemoryReport()

	if report == "" {
		t.Error("GetMemoryReport returned empty string")
	}

	// Check that report contains expected sections
	expectedStrings := []string{
		"Memory Report",
		"Heap Alloc",
		"Goroutines",
		"GC Cycles",
	}

	for _, s := range expectedStrings {
		if !contains(report, s) {
			t.Errorf("Report should contain '%s'", s)
		}
	}
}

func TestProfilerStartStop(t *testing.T) {
	p := NewProfiler(10, 100*time.Millisecond)

	p.Start()
	time.Sleep(250 * time.Millisecond)
	p.Stop()

	samples := p.GetSamples()
	if len(samples) < 1 {
		t.Error("Expected at least 1 sample after running profiler")
	}
}

func TestMemoryWatcher(t *testing.T) {
	var called atomic.Bool
	callback := func(stats MemoryStats) {
		called.Store(true)
	}

	// Set a very low threshold to trigger callback
	w := NewMemoryWatcher(1, callback, 100*time.Millisecond)
	w.Start()
	time.Sleep(200 * time.Millisecond)
	w.Stop()

	if !called.Load() {
		t.Error("Callback should have been called")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
