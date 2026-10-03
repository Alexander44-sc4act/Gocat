// Package network provides optimized connection pooling
package network

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// OptimizedPool provides high-performance connection pooling
type OptimizedPool struct {
	config      *OptimizedPoolConfig
	pools       sync.Map // map[string]*addressPool
	stats       OptimizedPoolStats
	closed      atomic.Bool
	dialFunc    DialFunc
	healthCheck HealthCheckFunc
}

// DialFunc is a function type for creating connections
type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

// HealthCheckFunc is a function type for checking connection health
type HealthCheckFunc func(conn net.Conn) bool

// OptimizedPoolConfig holds configuration for optimized pool
type OptimizedPoolConfig struct {
	MaxConnsPerHost     int
	MaxIdleConnsPerHost int
	MaxIdleTime         time.Duration
	MaxLifetime         time.Duration
	DialTimeout         time.Duration
	IdleCheckInterval   time.Duration
	EnablePrewarming    bool
	PrewarmHosts        []string
}

// OptimizedPoolStats holds pool statistics
type OptimizedPoolStats struct {
	TotalConns    int64
	ActiveConns   int64
	IdleConns     int64
	Hits          int64
	Misses        int64
	Timeouts      int64
	Errors        int64
	AvgWaitTime   int64 // nanoseconds
	TotalWaitTime int64
	WaitCount     int64
}

// addressPool manages connections for a single address
type addressPool struct {
	address     string
	idle        chan *pooledConn
	active      int64
	created     int64
	mu          sync.Mutex
	lastCleanup time.Time
}

// pooledConn wraps a connection with metadata
type pooledConn struct {
	conn      net.Conn
	createdAt time.Time
	lastUsed  time.Time
	useCount  int64
}

// DefaultOptimizedPoolConfig returns default configuration
func DefaultOptimizedPoolConfig() *OptimizedPoolConfig {
	return &OptimizedPoolConfig{
		MaxConnsPerHost:     100,
		MaxIdleConnsPerHost: 10,
		MaxIdleTime:         90 * time.Second,
		MaxLifetime:         30 * time.Minute,
		DialTimeout:         30 * time.Second,
		IdleCheckInterval:   30 * time.Second,
		EnablePrewarming:    false,
	}
}

// NewOptimizedPool creates a new optimized connection pool
func NewOptimizedPool(config *OptimizedPoolConfig, dialFunc DialFunc) *OptimizedPool {
	if config == nil {
		config = DefaultOptimizedPoolConfig()
	}
	if dialFunc == nil {
		dialFunc = defaultDialFunc
	}

	pool := &OptimizedPool{
		config:   config,
		dialFunc: dialFunc,
		healthCheck: func(conn net.Conn) bool {
			return conn != nil
		},
	}

	// Start background cleanup
	go pool.cleanupLoop()

	// Prewarm if enabled
	if config.EnablePrewarming && len(config.PrewarmHosts) > 0 {
		go pool.prewarm()
	}

	return pool
}

func defaultDialFunc(ctx context.Context, network, address string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, network, address)
}

// Get retrieves a connection from the pool
func (p *OptimizedPool) Get(ctx context.Context, address string) (net.Conn, error) {
	if p.closed.Load() {
		return nil, ErrPoolClosed
	}

	startTime := time.Now()
	defer func() {
		waitTime := time.Since(startTime).Nanoseconds()
		atomic.AddInt64(&p.stats.TotalWaitTime, waitTime)
		atomic.AddInt64(&p.stats.WaitCount, 1)
	}()

	// Get or create address pool
	ap := p.getAddressPool(address)

	// Try to get idle connection
	select {
	case pc := <-ap.idle:
		if p.isValid(pc) {
			atomic.AddInt64(&p.stats.Hits, 1)
			atomic.AddInt64(&p.stats.ActiveConns, 1)
			atomic.AddInt64(&p.stats.IdleConns, -1)
			atomic.AddInt64(&ap.active, 1)
			pc.lastUsed = time.Now()
			pc.useCount++
			return &pooledConnWrapper{pc: pc, pool: p, address: address}, nil
		}
		// Connection invalid, close it
		pc.conn.Close()
	default:
		// No idle connection available
	}

	atomic.AddInt64(&p.stats.Misses, 1)

	// Check if we can create new connection
	if atomic.LoadInt64(&ap.created) >= int64(p.config.MaxConnsPerHost) {
		// Wait for available connection
		select {
		case pc := <-ap.idle:
			if p.isValid(pc) {
				atomic.AddInt64(&p.stats.Hits, 1)
				atomic.AddInt64(&p.stats.ActiveConns, 1)
				atomic.AddInt64(&p.stats.IdleConns, -1)
				atomic.AddInt64(&ap.active, 1)
				pc.lastUsed = time.Now()
				pc.useCount++
				return &pooledConnWrapper{pc: pc, pool: p, address: address}, nil
			}
			pc.conn.Close()
		case <-ctx.Done():
			atomic.AddInt64(&p.stats.Timeouts, 1)
			return nil, ctx.Err()
		}
	}

	// Create new connection
	dialCtx, cancel := context.WithTimeout(ctx, p.config.DialTimeout)
	defer cancel()

	conn, err := p.dialFunc(dialCtx, "tcp", address)
	if err != nil {
		atomic.AddInt64(&p.stats.Errors, 1)
		return nil, err
	}

	atomic.AddInt64(&ap.created, 1)
	atomic.AddInt64(&p.stats.TotalConns, 1)
	atomic.AddInt64(&p.stats.ActiveConns, 1)
	atomic.AddInt64(&ap.active, 1)

	pc := &pooledConn{
		conn:      conn,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		useCount:  1,
	}

	return &pooledConnWrapper{pc: pc, pool: p, address: address}, nil
}

// Put returns a connection to the pool
func (p *OptimizedPool) Put(address string, pc *pooledConn) {
	if p.closed.Load() || !p.isValid(pc) {
		pc.conn.Close()
		return
	}

	ap := p.getAddressPool(address)
	atomic.AddInt64(&ap.active, -1)
	atomic.AddInt64(&p.stats.ActiveConns, -1)

	// Try to return to idle pool
	select {
	case ap.idle <- pc:
		atomic.AddInt64(&p.stats.IdleConns, 1)
	default:
		// Pool full, close connection
		pc.conn.Close()
		atomic.AddInt64(&ap.created, -1)
		atomic.AddInt64(&p.stats.TotalConns, -1)
	}
}

func (p *OptimizedPool) getAddressPool(address string) *addressPool {
	if ap, ok := p.pools.Load(address); ok {
		return ap.(*addressPool)
	}

	ap := &addressPool{
		address:     address,
		idle:        make(chan *pooledConn, p.config.MaxIdleConnsPerHost),
		lastCleanup: time.Now(),
	}

	actual, _ := p.pools.LoadOrStore(address, ap)
	return actual.(*addressPool)
}

func (p *OptimizedPool) isValid(pc *pooledConn) bool {
	if pc == nil || pc.conn == nil {
		return false
	}

	now := time.Now()

	// Check lifetime
	if now.Sub(pc.createdAt) > p.config.MaxLifetime {
		return false
	}

	// Check idle time
	if now.Sub(pc.lastUsed) > p.config.MaxIdleTime {
		return false
	}

	// Health check
	if p.healthCheck != nil && !p.healthCheck(pc.conn) {
		return false
	}

	return true
}

func (p *OptimizedPool) cleanupLoop() {
	ticker := time.NewTicker(p.config.IdleCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		if p.closed.Load() {
			return
		}
		p.cleanup()
	}
}

func (p *OptimizedPool) cleanup() {
	p.pools.Range(func(key, value interface{}) bool {
		ap := value.(*addressPool)
		p.cleanupAddressPool(ap)
		return true
	})
}

func (p *OptimizedPool) cleanupAddressPool(ap *addressPool) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	// Drain and check idle connections
	validConns := make([]*pooledConn, 0)

	for {
		select {
		case pc := <-ap.idle:
			if p.isValid(pc) {
				validConns = append(validConns, pc)
			} else {
				pc.conn.Close()
				atomic.AddInt64(&ap.created, -1)
				atomic.AddInt64(&p.stats.TotalConns, -1)
				atomic.AddInt64(&p.stats.IdleConns, -1)
			}
		default:
			goto done
		}
	}

done:
	// Return valid connections
	for _, pc := range validConns {
		select {
		case ap.idle <- pc:
		default:
			pc.conn.Close()
			atomic.AddInt64(&ap.created, -1)
			atomic.AddInt64(&p.stats.TotalConns, -1)
			atomic.AddInt64(&p.stats.IdleConns, -1)
		}
	}

	ap.lastCleanup = time.Now()
}

func (p *OptimizedPool) prewarm() {
	ctx := context.Background()
	for _, host := range p.config.PrewarmHosts {
		for i := 0; i < p.config.MaxIdleConnsPerHost/2; i++ {
			conn, err := p.Get(ctx, host)
			if err == nil {
				conn.Close() // Returns to pool
			}
		}
	}
}

// Stats returns pool statistics
func (p *OptimizedPool) Stats() OptimizedPoolStats {
	stats := OptimizedPoolStats{
		TotalConns:    atomic.LoadInt64(&p.stats.TotalConns),
		ActiveConns:   atomic.LoadInt64(&p.stats.ActiveConns),
		IdleConns:     atomic.LoadInt64(&p.stats.IdleConns),
		Hits:          atomic.LoadInt64(&p.stats.Hits),
		Misses:        atomic.LoadInt64(&p.stats.Misses),
		Timeouts:      atomic.LoadInt64(&p.stats.Timeouts),
		Errors:        atomic.LoadInt64(&p.stats.Errors),
		TotalWaitTime: atomic.LoadInt64(&p.stats.TotalWaitTime),
		WaitCount:     atomic.LoadInt64(&p.stats.WaitCount),
	}

	if stats.WaitCount > 0 {
		stats.AvgWaitTime = stats.TotalWaitTime / stats.WaitCount
	}

	return stats
}

// Close closes the pool and all connections
func (p *OptimizedPool) Close() error {
	if p.closed.Swap(true) {
		return nil
	}

	p.pools.Range(func(key, value interface{}) bool {
		ap := value.(*addressPool)
		close(ap.idle)
		for pc := range ap.idle {
			pc.conn.Close()
		}
		return true
	})

	return nil
}

// pooledConnWrapper wraps pooledConn to implement net.Conn
type pooledConnWrapper struct {
	pc      *pooledConn
	pool    *OptimizedPool
	address string
	closed  bool
}

func (w *pooledConnWrapper) Read(b []byte) (n int, err error) {
	return w.pc.conn.Read(b)
}

func (w *pooledConnWrapper) Write(b []byte) (n int, err error) {
	return w.pc.conn.Write(b)
}

func (w *pooledConnWrapper) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	w.pool.Put(w.address, w.pc)
	return nil
}

func (w *pooledConnWrapper) LocalAddr() net.Addr {
	return w.pc.conn.LocalAddr()
}

func (w *pooledConnWrapper) RemoteAddr() net.Addr {
	return w.pc.conn.RemoteAddr()
}

func (w *pooledConnWrapper) SetDeadline(t time.Time) error {
	return w.pc.conn.SetDeadline(t)
}

func (w *pooledConnWrapper) SetReadDeadline(t time.Time) error {
	return w.pc.conn.SetReadDeadline(t)
}

func (w *pooledConnWrapper) SetWriteDeadline(t time.Time) error {
	return w.pc.conn.SetWriteDeadline(t)
}

// ErrPoolClosed is returned when pool is closed
var ErrPoolClosed = &poolError{msg: "connection pool is closed"}

type poolError struct {
	msg string
}

func (e *poolError) Error() string {
	return e.msg
}
