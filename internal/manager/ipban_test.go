package manager

import (
	"testing"
	"time"
)

func TestIPBanManager_IsBanned(t *testing.T) {
	manager := NewIPBanManager(3, 5*time.Second, []string{"192.168.1.1"})
	defer manager.Stop()

	// Test non-banned IP
	if manager.IsBanned("10.0.0.1") {
		t.Error("IP should not be banned initially")
	}

	// Test whitelisted IP
	if manager.IsBanned("192.168.1.1") {
		t.Error("Whitelisted IP should never be banned")
	}
}

func TestIPBanManager_RecordFailure(t *testing.T) {
	manager := NewIPBanManager(3, 1*time.Second, []string{})
	defer manager.Stop()

	ip := "10.0.0.1"

	// Record failures
	for i := 0; i < 2; i++ {
		manager.RecordFailure(ip)
		if manager.IsBanned(ip) {
			t.Errorf("IP should not be banned after %d failures", i+1)
		}
	}

	// Third failure should trigger ban
	manager.RecordFailure(ip)
	if !manager.IsBanned(ip) {
		t.Error("IP should be banned after 3 failures")
	}

	// Ban should expire
	time.Sleep(1500 * time.Millisecond)
	if manager.IsBanned(ip) {
		t.Error("IP should not be banned after expiry")
	}
}

func TestIPBanManager_RecordSuccess(t *testing.T) {
	manager := NewIPBanManager(3, 5*time.Second, []string{})
	defer manager.Stop()

	ip := "10.0.0.1"

	// Record failures
	manager.RecordFailure(ip)
	manager.RecordFailure(ip)

	// Success should reset counter
	manager.RecordSuccess(ip)

	// These failures should not trigger ban yet
	manager.RecordFailure(ip)
	manager.RecordFailure(ip)
	if manager.IsBanned(ip) {
		t.Error("IP should not be banned after success reset")
	}
}

func TestIPBanManager_UnbanIP(t *testing.T) {
	manager := NewIPBanManager(3, 5*time.Second, []string{})
	defer manager.Stop()

	ip := "10.0.0.1"

	// Trigger ban
	for i := 0; i < 3; i++ {
		manager.RecordFailure(ip)
	}

	if !manager.IsBanned(ip) {
		t.Fatal("IP should be banned")
	}

	// Manually unban
	manager.UnbanIP(ip)
	if manager.IsBanned(ip) {
		t.Error("IP should not be banned after manual unban")
	}
}

func TestIPBanManager_GetBannedIPs(t *testing.T) {
	manager := NewIPBanManager(2, 5*time.Second, []string{})
	defer manager.Stop()

	// Ban multiple IPs
	ips := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	for _, ip := range ips {
		for i := 0; i < 2; i++ {
			manager.RecordFailure(ip)
		}
	}

	bannedIPs := manager.GetBannedIPs()
	if len(bannedIPs) != 3 {
		t.Errorf("Expected 3 banned IPs, got %d", len(bannedIPs))
	}
}

func TestIPBanManager_Whitelist(t *testing.T) {
	whitelist := []string{"192.168.1.1", "192.168.1.2"}
	manager := NewIPBanManager(2, 5*time.Second, whitelist)
	defer manager.Stop()

	// Try to ban whitelisted IPs
	for _, ip := range whitelist {
		for i := 0; i < 5; i++ {
			manager.RecordFailure(ip)
		}
		if manager.IsBanned(ip) {
			t.Errorf("Whitelisted IP %s should never be banned", ip)
		}
	}
}

// Benchmark tests
func BenchmarkIPBanManager_IsBanned(b *testing.B) {
	manager := NewIPBanManager(3, 5*time.Second, []string{})
	defer manager.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.IsBanned("10.0.0.1")
	}
}

func BenchmarkIPBanManager_RecordFailure(b *testing.B) {
	manager := NewIPBanManager(3, 5*time.Second, []string{})
	defer manager.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.RecordFailure("10.0.0.1")
	}
}

func BenchmarkIPBanManager_RecordSuccess(b *testing.B) {
	manager := NewIPBanManager(3, 5*time.Second, []string{})
	defer manager.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.RecordSuccess("10.0.0.1")
	}
}
