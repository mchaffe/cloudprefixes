package db

import (
	"net"
	"reflect"
	"testing"

	_ "modernc.org/sqlite"
)

func stringPointer(s string) *string {
	return &s
}

func TestPrefixManager_AddPrefix(t *testing.T) {
	// Create a temporary in-memory database for testing
	manager, err := NewPrefixManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create IPRangeManager: %v", err)
	}
	defer manager.Close()

	// Initialize test cases
	tests := []struct {
		name    string
		info    PrefixInfo
		wantErr bool
	}{
		{
			name: "Valid IPv4 CIDR",
			info: PrefixInfo{
				Prefix:   "192.168.1.0/24",
				Platform: "AWS",
				Region:   stringPointer("us-east-1"),
				Service:  stringPointer("EC2"),
			},
			wantErr: false,
		},
		{
			name: "Valid IPv6 CIDR",
			info: PrefixInfo{
				Prefix:   "2001:db8::/32",
				Platform: "GCP",
				Region:   stringPointer("global"),
				Service:  stringPointer("Compute"),
			},
			wantErr: false,
		},
		{
			name: "Invalid CIDR",
			info: PrefixInfo{
				Prefix:   "invalid_cidr",
				Platform: "AWS",
				Region:   stringPointer("us-west-2"),
				Service:  stringPointer("S3"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := manager.AddPrefix(tt.info); (err != nil) != tt.wantErr {
				t.Errorf("PrefixManager.AddPrefix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrefixManager_AddPrefixBatch(t *testing.T) {
	manager, err := NewPrefixManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create IPRangeManager: %v", err)
	}
	defer manager.Close()

	tests := []struct {
		name    string
		infos   []PrefixInfo
		wantErr bool
	}{
		{
			name: "Valid Batch of Prefixes",
			infos: []PrefixInfo{
				{
					Prefix:   "192.168.2.0/24",
					Platform: "AWS",
					Region:   stringPointer("us-east-1"),
					Service:  stringPointer("EC2"),
				},
				{
					Prefix:   "2001:db8::/32",
					Platform: "Azure",
					Region:   stringPointer("global"),
					Service:  stringPointer("VM"),
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid CIDR in Batch",
			infos: []PrefixInfo{
				{
					Prefix:   "invalid_cidr",
					Platform: "GCP",
					Region:   stringPointer("us-central1"),
					Service:  stringPointer("Kubernetes"),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := manager.AddPrefixBatch(tt.infos); (err != nil) != tt.wantErr {
				t.Errorf("PrefixManager.AddPrefixBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrefixManager_ContainsIP(t *testing.T) {
	manager, err := NewPrefixManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create IPRangeManager: %v", err)
	}
	defer manager.Close()

	// Pre-populate database with known prefixes
	err = manager.AddPrefix(PrefixInfo{
		Prefix:   "192.168.3.0/24",
		Platform: "AWS",
		Region:   stringPointer("us-east-1"),
		Service:  stringPointer("EC2"),
	})
	if err != nil {
		t.Fatalf("Failed to add prefix: %v", err)
	}

	err = manager.AddPrefix(PrefixInfo{
		Prefix:   "2001:db8::/32",
		Platform: "Azure",
		Region:   stringPointer("global"),
		Service:  stringPointer("VM"),
	})
	if err != nil {
		t.Fatalf("Failed to add prefix: %v", err)
	}

	tests := []struct {
		name     string
		ip       string
		want     bool
		wantErr  bool
		expected int // Expected number of prefixes in result
	}{
		{"IP inside IPv4 range", "192.168.3.10", true, false, 1},
		{"IP inside IPv6 range", "2001:db8::1", true, false, 1},
		{"IP outside range", "203.0.113.5", false, false, 0},
		{"Invalid IP", "invalid_ip", false, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, results, err := manager.ContainsIP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrefixManager.ContainsIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PrefixManager.ContainsIP() = %v, want %v", got, tt.want)
			}
			if len(results) != tt.expected {
				t.Errorf("Expected %d prefixes, but got %d", tt.expected, len(results))
			}
		})
	}
}

func Test_ipToInts(t *testing.T) {
	tests := []struct {
		name     string
		ip       net.IP
		wantHigh uint64
		wantLow  uint64
		wantErr  bool
	}{
		{"IPv4", net.ParseIP("203.0.113.1").To4(), 0, 3405803777, false},
		{"IPv6", net.ParseIP("2001:db8::1").To16(), 2306139568115548160, 2306139568115548160, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHigh, gotLow, err := ipToInts(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ipToInts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHigh != tt.wantHigh {
				t.Errorf("ipToInts() gotHigh = %v, want %v", gotHigh, tt.wantHigh)
			}
			if gotLow != tt.wantLow {
				t.Errorf("ipToInts() gotLow = %v, want %v", gotLow, tt.wantLow)
			}
		})
	}
}

func Test_lastIP(t *testing.T) {
	tests := []struct {
		name  string
		ipNet *net.IPNet
		want  net.IP
	}{
		{
			"IPv4",
			&net.IPNet{IP: net.IPv4(203, 0, 113, 0).To4(), Mask: net.CIDRMask(24, 32)},
			net.IPv4(203, 0, 113, 255).To4(),
		},
		{
			"IPv6",
			&net.IPNet{IP: net.ParseIP("2001:db8::").To16(), Mask: net.CIDRMask(64, 128)},
			net.ParseIP("2001:db8::ffff:ffff:ffff:ffff").To16(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lastIP(tt.ipNet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lastIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrefixManager_ClearAllData(t *testing.T) {
	manager, err := NewPrefixManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create IPRangeManager: %v", err)
	}
	defer manager.Close()

	// Pre-populate database with known prefixes
	err = manager.AddPrefix(PrefixInfo{
		Prefix:   "192.168.4.0/24",
		Platform: "AWS",
		Region:   stringPointer("us-east-1"),
		Service:  stringPointer("EC2"),
	})
	if err != nil {
		t.Fatalf("Failed to add prefix: %v", err)
	}

	tests := []struct {
		name    string
		m       *PrefixManager
		wantErr bool
	}{
		{
			name:    "Clear Data",
			m:       manager,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.ClearAllData(); (err != nil) != tt.wantErr {
				t.Errorf("PrefixManager.ClearAllData() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Ensure the data is cleared
			_, results, err := tt.m.ContainsIP("192.168.4.10")
			if err != nil {
				t.Fatalf("Failed to check IP after clear: %v", err)
			}
			if len(results) != 0 {
				t.Errorf("Expected no prefixes after clearing data, but found %d", len(results))
			}
		})
	}
}
