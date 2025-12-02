package manager

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// BanRecord represents a single IP ban record for persistence
type BanRecord struct {
	IP        string    `json:"ip"`
	BannedAt  time.Time `json:"banned_at"`
	ExpiresAt time.Time `json:"expires_at"`
	FailCount int       `json:"fail_count"`
}

// IPBanManager manages IP banning based on authentication failures
type IPBanManager struct {
	mu              sync.RWMutex
	bannedIPs       map[string]time.Time // IP -> ban expiry time
	bannedFailCount map[string]int       // IP -> failure count at time of ban
	failureCounts   map[string]int       // IP -> current failure count
	maxFailures     int
	banDuration     time.Duration
	whitelist       map[string]bool
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
	persistFile     string // Path to persistence file
}

// NewIPBanManager creates a new IP ban manager
func NewIPBanManager(maxFailures int, banDuration time.Duration, whitelist []string) *IPBanManager {
	wl := make(map[string]bool)
	for _, ip := range whitelist {
		wl[ip] = true
	}

	manager := &IPBanManager{
		bannedIPs:       make(map[string]time.Time),
		bannedFailCount: make(map[string]int),
		failureCounts:   make(map[string]int),
		maxFailures:     maxFailures,
		banDuration:     banDuration,
		whitelist:       wl,
		cleanupInterval: time.Minute,
		stopCleanup:     make(chan struct{}),
		persistFile:     "data/ipban.json", // Default persistence file
	}

	// Load persisted data
	manager.loadFromFile()

	// Start cleanup routine
	go manager.cleanupExpiredBans()

	return manager
}

// IsBanned checks if an IP is currently banned
func (m *IPBanManager) IsBanned(ip string) bool {
	// Whitelisted IPs are never banned
	if m.whitelist[ip] {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	expiry, exists := m.bannedIPs[ip]
	if !exists {
		return false
	}

	// Check if ban has expired
	if time.Now().After(expiry) {
		return false
	}

	return true
}

// RecordFailure records an authentication failure for an IP
func (m *IPBanManager) RecordFailure(ip string) {
	// Don't track whitelisted IPs
	if m.whitelist[ip] {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.failureCounts[ip]++

	// Ban the IP if it exceeds the threshold
	if m.failureCounts[ip] >= m.maxFailures {
		// Save the failure count that triggered the ban
		m.bannedFailCount[ip] = m.failureCounts[ip]
		m.bannedIPs[ip] = time.Now().Add(m.banDuration)
		// Reset failure count after banning
		delete(m.failureCounts, ip)

		// Persist the ban
		go m.saveToFile()
	}
}

// RecordSuccess records a successful authentication for an IP
func (m *IPBanManager) RecordSuccess(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset failure count on success
	delete(m.failureCounts, ip)
}

// UnbanIP manually unbans an IP
func (m *IPBanManager) UnbanIP(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.bannedIPs, ip)
	delete(m.bannedFailCount, ip)
	delete(m.failureCounts, ip)

	// Persist the change
	go m.saveToFile()
}

// GetBannedIPs returns a list of currently banned IPs
func (m *IPBanManager) GetBannedIPs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	var banned []string
	for ip, expiry := range m.bannedIPs {
		if now.Before(expiry) {
			banned = append(banned, ip)
		}
	}
	return banned
}

// GetFailureCount returns the current failure count for an IP
func (m *IPBanManager) GetFailureCount(ip string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.failureCounts[ip]
}

// cleanupExpiredBans periodically removes expired bans
func (m *IPBanManager) cleanupExpiredBans() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			changed := false
			for ip, expiry := range m.bannedIPs {
				if now.After(expiry) {
					delete(m.bannedIPs, ip)
					changed = true
				}
			}
			m.mu.Unlock()

			// Persist if anything changed
			if changed {
				go m.saveToFile()
			}
		case <-m.stopCleanup:
			return
		}
	}
}

// Stop stops the cleanup routine and saves final state
func (m *IPBanManager) Stop() {
	close(m.stopCleanup)
	m.saveToFile() // Save final state before stopping
}

// saveToFile persists the current ban state to disk
func (m *IPBanManager) saveToFile() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create data directory if it doesn't exist
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	// Prepare records
	var records []BanRecord
	now := time.Now()
	for ip, expiry := range m.bannedIPs {
		// Only save non-expired bans
		if now.Before(expiry) {
			record := BanRecord{
				IP:        ip,
				ExpiresAt: expiry,
				BannedAt:  expiry.Add(-m.banDuration),
			}
			// Add the failure count that triggered the ban
			if failCount, exists := m.bannedFailCount[ip]; exists {
				record.FailCount = failCount
			}
			records = append(records, record)
		}
	}

	// Add IPs with failure counts that haven't been banned yet
	for ip, count := range m.failureCounts {
		// Check if this IP already has a ban record
		found := false
		for i := range records {
			if records[i].IP == ip {
				found = true
				break
			}
		}
		// If not banned but has failures, add it
		if !found && count > 0 {
			records = append(records, BanRecord{
				IP:        ip,
				FailCount: count,
			})
		}
	}

	// Write to file
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.persistFile, data, 0644)
}

// loadFromFile loads the ban state from disk
func (m *IPBanManager) loadFromFile() error {
	data, err := os.ReadFile(m.persistFile)
	if err != nil {
		// File doesn't exist is not an error on first run
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var records []BanRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	// Restore bans and failure counts
	now := time.Now()
	for _, record := range records {
		// Only restore non-expired bans
		if !record.ExpiresAt.IsZero() && now.Before(record.ExpiresAt) {
			m.bannedIPs[record.IP] = record.ExpiresAt
			// Restore the failure count that triggered the ban
			if record.FailCount > 0 {
				m.bannedFailCount[record.IP] = record.FailCount
			}
		} else if record.FailCount > 0 {
			// If not banned anymore（expired) but has failure count, restore it
			m.failureCounts[record.IP] = record.FailCount
		}
	}

	return nil
}
