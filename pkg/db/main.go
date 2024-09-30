package db

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"math/big"
	"net"

	_ "github.com/mattn/go-sqlite3"
)

type PrefixInfo struct {
	Prefix   string  `json:"prefix"`
	Platform string  `json:"platform"`
	Region   *string `json:"region,omitempty"`
	Service  *string `json:"service,omitempty"`
	Metadata *string `json:"metadata,omitempty"`
}

type PrefixManager struct {
	db *sql.DB
}

func NewIPRangeManager(dbPath string) (*PrefixManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	manager := &PrefixManager{db: db}
	err = manager.initDB()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error initializing database: %v", err)
	}

	return manager, nil
}

func (m *PrefixManager) initDB() error {
	_, err := m.db.Exec(`
        CREATE TABLE IF NOT EXISTS cloud_prefixes (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
			service TEXT,
			platform TEXT,
			region TEXT,
            prefix TEXT,
			start_ip_high INTEGER,
            start_ip_low INTEGER,
            end_ip_high INTEGER,
            end_ip_low INTEGER,
			ip_version INTEGER,
			metadata JSONB
        )
    `)
	return err
}

func (m *PrefixManager) AddPrefix(info PrefixInfo) error {
	_, ipNet, err := net.ParseCIDR(info.Prefix)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %v", err)
	}

	startIPHigh, startIPLow, err := ipToInts(ipNet.IP)
	if err != nil {
		return err
	}
	endIPHigh, endIPLow, err := ipToInts(lastIP(ipNet))
	if err != nil {
		return err
	}
	ipVersion := 4
	if ipNet.IP.To4() == nil {
		ipVersion = 6
	}

	_, err = m.db.Exec(`
        INSERT OR REPLACE INTO cloud_prefixes 
        (prefix, start_ip_high, start_ip_low, end_ip_high, end_ip_low, ip_version, region, platform, service, metadata) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		info.Prefix, startIPHigh, startIPLow, endIPHigh, endIPLow, ipVersion, info.Region, info.Platform, info.Service, info.Metadata)
	return err
}

func (m *PrefixManager) AddPrefixBatch(infos []PrefixInfo) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
        INSERT OR REPLACE INTO cloud_prefixes 
        (prefix, start_ip_high, start_ip_low, end_ip_high, end_ip_low, ip_version, region, platform, service, metadata) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, info := range infos {
		_, ipNet, err := net.ParseCIDR(info.Prefix)
		if err != nil {
			return fmt.Errorf("invalid CIDR %s: %v", info.Prefix, err)
		}

		startIPHigh, startIPLow, err := ipToInts(ipNet.IP)
		if err != nil {
			return err
		}
		endIPHigh, endIPLow, err := ipToInts(lastIP(ipNet))
		if err != nil {
			return err
		}
		ipVersion := 4
		if ipNet.IP.To4() == nil {
			ipVersion = 6
		}

		_, err = stmt.Exec(info.Prefix, startIPHigh, startIPLow, endIPHigh, endIPLow, ipVersion, info.Region, info.Platform, info.Service, info.Metadata)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (m *PrefixManager) ContainsIP(ip string) (bool, []PrefixInfo, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false, []PrefixInfo{}, fmt.Errorf("invalid IP address")
	}

	ipHigh, ipLow, err := ipToInts(parsedIP)
	if err != nil {
		return false, []PrefixInfo{}, err
	}

	ipVersion := 4
	if parsedIP.To4() == nil {
		ipVersion = 6
	}

	rows, err := m.db.Query(`
        SELECT prefix, region, platform, service, metadata 
        FROM cloud_prefixes
        WHERE start_ip_high <= ? AND start_ip_low <= ? 
        AND end_ip_high >= ? AND end_ip_low >= ? 
        AND ip_version = ?`,
		ipHigh, ipLow, ipHigh, ipLow, ipVersion)
	if err != nil {
		return false, []PrefixInfo{}, err
	}

	var results []PrefixInfo
	for rows.Next() {
		var info PrefixInfo
		if err := rows.Scan(&info.Prefix, &info.Region, &info.Platform, &info.Service, &info.Metadata); err != nil {
			return false, []PrefixInfo{}, err
		}
		results = append(results, info)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return false, []PrefixInfo{}, err
	}

	// Check if no results were found
	if len(results) == 0 {
		return false, []PrefixInfo{}, nil
	}

	return true, results, nil
}

func (m *PrefixManager) Close() error {
	return m.db.Close()
}

// SQLite doesn't have 128bit integers or dedicated IP type. So in order to store
// them in a way to make querying efficient, split them into two 64bit integers
func ipToInts(ip net.IP) (high uint64, low uint64, err error) {
	if ip == nil {
		return 0, 0, fmt.Errorf("nil IP address")
	}

	ipv4 := ip.To4()
	if ipv4 != nil {
		// Handle IPv4
		return 0, uint64(binary.BigEndian.Uint32(ipv4)), nil
	}

	// Handle IPv6
	ipv6 := ip.To16()
	if ipv6 == nil {
		return 0, 0, fmt.Errorf("invalid IP address")
	}

	ipInt := new(big.Int).SetBytes(ipv6)
	high = ipInt.Rsh(ipInt, 64).Uint64()
	low = ipInt.Uint64()

	return high, low, nil
}

func lastIP(ipNet *net.IPNet) net.IP {
	lastIP := make(net.IP, len(ipNet.IP))
	copy(lastIP, ipNet.IP)
	for i := range lastIP {
		lastIP[i] |= ^ipNet.Mask[i]
	}
	return lastIP
}

func (m *PrefixManager) ClearAllData() error {
	_, err := m.db.Exec("DELETE FROM cloud_prefixes")
	if err != nil {
		return fmt.Errorf("failed to clear database: %v", err)
	}
	return nil
}
