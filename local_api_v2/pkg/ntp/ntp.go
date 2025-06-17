package ntp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

// TimeService provides NTP time synchronization with Moscow timezone support
type TimeService struct {
	servers     []string
	timeout     time.Duration
	syncedTime  time.Time
	localTime   time.Time
	offset      time.Duration
	lastSync    time.Time
	mu          sync.RWMutex
	isConnected bool
	enabled     bool
	stopChan    chan struct{}
	moscowTZ    *time.Location
}

// Config holds NTP service configuration
type Config struct {
	Enabled      bool
	Servers      []string
	Timeout      time.Duration
	SyncInterval time.Duration
}

// NewTimeService creates a new NTP time service with configuration
func NewTimeService(config Config) *TimeService {
	servers := config.Servers
	if len(servers) == 0 {
		servers = []string{
			"ru.pool.ntp.org",     // Russia NTP pool
			"europe.pool.ntp.org", // Europe NTP pool
			"0.ru.pool.ntp.org",   // Russia specific
			"1.ru.pool.ntp.org",   // Russia specific
			"pool.ntp.org",        // Global fallback
		}
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// Load Moscow timezone
	moscowTZ, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("Warning: Could not load Moscow timezone: %v", err)
		moscowTZ = time.UTC
	}

	return &TimeService{
		servers:  servers,
		timeout:  timeout,
		enabled:  config.Enabled,
		stopChan: make(chan struct{}),
		moscowTZ: moscowTZ,
	}
}

// Start begins the NTP service with periodic synchronization
func (ts *TimeService) Start(ctx context.Context, syncInterval time.Duration) error {
	if !ts.enabled {
		log.Println("NTP service disabled - using system time")
		return nil
	}

	log.Printf("Starting NTP service with servers: %v", ts.servers)

	// Initial sync
	if err := ts.Sync(); err != nil {
		log.Printf("Initial NTP sync failed: %v", err)
		// Don't return error - continue with periodic attempts
	}

	// Start periodic sync in background
	go func() {
		ticker := time.NewTicker(syncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("NTP service stopped by context")
				return
			case <-ts.stopChan:
				log.Println("NTP service stopped")
				return
			case <-ticker.C:
				if err := ts.Sync(); err != nil {
					log.Printf("Periodic NTP sync failed: %v", err)
				}
			}
		}
	}()

	return nil
}

// Stop stops the NTP service
func (ts *TimeService) Stop() {
	if ts.stopChan != nil {
		close(ts.stopChan)
	}
}

// Sync synchronizes with NTP servers
func (ts *TimeService) Sync() error {
	var lastErr error

	for _, server := range ts.servers {
		response, err := ntp.QueryWithOptions(server, ntp.QueryOptions{
			Timeout: ts.timeout,
		})

		if err != nil {
			lastErr = err
			log.Printf("NTP sync failed for %s: %v", server, err)
			continue
		}

		ts.mu.Lock()
		ts.syncedTime = response.Time
		ts.localTime = time.Now()
		ts.offset = response.ClockOffset
		ts.lastSync = time.Now()
		ts.isConnected = true
		ts.mu.Unlock()

		log.Printf("✅ NTP synced with %s (offset: %v) Local time: %v", server, response.ClockOffset, ts.localTime)
		return nil
	}

	ts.mu.Lock()
	ts.isConnected = false
	ts.mu.Unlock()

	return fmt.Errorf("failed to sync with any NTP server, last error: %v", lastErr)
}

// Now returns the current time, using NTP if available
func (ts *TimeService) Now() time.Time {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if !ts.isConnected || ts.lastSync.IsZero() {
		// Fallback to local time if NTP is not available
		return time.Now()
	}

	// Calculate current time based on NTP sync
	elapsed := time.Since(ts.localTime)
	return ts.syncedTime.Add(elapsed)
}

// NowInMoscow returns current time in Moscow timezone
func (ts *TimeService) NowInMoscow() time.Time {
	currentTime := ts.Now()

	return currentTime.In(ts.moscowTZ).Add(3 * time.Hour)
}

// Unix returns Unix timestamp using NTP time
func (ts *TimeService) Unix() int64 {
	return ts.Now().Unix()
}

// IsConnected returns true if NTP is working
func (ts *TimeService) IsConnected() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.isConnected
}

// GetOffset returns the time offset from NTP
func (ts *TimeService) GetOffset() time.Duration {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.offset
}

func (ts *TimeService) IsEnabled() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.enabled
}

// StartPeriodicSync starts automatic NTP synchronization
func (ts *TimeService) StartPeriodicSync(interval time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		// Initial sync
		if err := ts.Sync(); err != nil {
			log.Printf("Initial NTP sync failed: %v", err)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ts.Sync(); err != nil {
					log.Printf("Periodic NTP sync failed: %v", err)
				}
			case <-stop:
				log.Println("NTP periodic sync stopped")
				return
			}
		}
	}()

	return stop
}
