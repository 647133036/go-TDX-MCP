package tdx

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bensema/gotdx"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// DefaultSections are the connect.cfg sections that hold quote servers on the
// standard TDX protocol (port 7709). [HFHost] carries the high-frequency
// redundant entries, which are not duplicates of [HQHOST].
var DefaultSections = []string{"HQHOST", "HFHost"}

// ParseConnectCfgHosts parses a TDX client connect.cfg file and extracts the
// server entries from the given section (e.g. "HQHOST" for main quote hosts).
//
// connect.cfg uses the INI format with GBK encoding. Each entry is a triplet:
//
//	HostNameNN=<name>   (GBK encoded)
//	IPAddressNN=<ip>
//	PortNN=<port>
//
// section is matched case-insensitively. Entries are returned in index order.
func ParseConnectCfgHosts(path string, section string) ([]gotdx.HostInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read connect.cfg: %w", err)
	}
	return ParseConnectCfgHostsBytes(data, section)
}

// LoadConnectCfgHosts loads hosts from the bundled official connect.cfg, or
// from path when it is non-empty (useful for pointing at a freshly downloaded
// client config). Sections defaults to DefaultSections.
func LoadConnectCfgHosts(path string, sections []string) ([]gotdx.HostInfo, error) {
	if len(sections) == 0 {
		sections = DefaultSections
	}
	data := embeddedConnectCfg
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read connect.cfg %s: %w", path, err)
		}
		data = b
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("no bundled connect.cfg available")
	}
	var out []gotdx.HostInfo
	for _, section := range sections {
		hosts, err := ParseConnectCfgHostsBytes(data, section)
		if err != nil {
			continue
		}
		out = append(out, hosts...)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no hosts found in sections %v", sections)
	}
	return out, nil
}

// MainHostsWithConnectCfg returns gotdx's built-in hosts merged with the
// official connect.cfg servers (extra first, so freshly published official
// servers are probed and used before the stale built-in list). This is the
// drop-in replacement for gotdx.MainHosts() that makes any probing path use
// the same servers the desktop client does.
func MainHostsWithConnectCfg() []gotdx.HostInfo {
	extra, err := LoadConnectCfgHosts("", nil)
	if err != nil || len(extra) == 0 {
		return gotdx.MainHosts()
	}
	base := gotdx.MainHosts()
	out := make([]gotdx.HostInfo, 0, len(extra)+len(base))
	seen := make(map[string]bool, len(extra)+len(base))
	for _, h := range append(extra, base...) {
		addr := h.Address()
		if addr == "" || seen[addr] {
			continue
		}
		seen[addr] = true
		out = append(out, h)
	}
	return out
}

// EmbeddedConnectCfgHosts parses the bundled connect.cfg, handy for tests and
// for reporting which official servers the collector probes.
func EmbeddedConnectCfgHosts() ([]gotdx.HostInfo, error) {
	return LoadConnectCfgHosts("", DefaultSections)
}

// ParseConnectCfgHostsBytes is ParseConnectCfgHosts for in-memory config data,
// so the bundled official connect.cfg can be used without touching the disk.
func ParseConnectCfgHostsBytes(data []byte, section string) ([]gotdx.HostInfo, error) {
	utf8, err := simplifiedchinese.GBK.NewDecoder().Bytes(data)
	if err != nil {
		utf8 = data // fall back to raw bytes if decoding fails
	}
	lines := strings.Split(string(utf8), "\n")

	target := "[" + strings.ToUpper(section) + "]"
	inSection := false
	names := map[int]string{}
	ips := map[int]string{}
	ports := map[int]string{}

	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inSection = strings.ToUpper(line) == target
			continue
		}
		if !inSection {
			continue
		}
		eq := strings.Index(line, "=")
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		upper := strings.ToUpper(key)
		switch {
		case strings.HasPrefix(upper, "HOSTNAME"):
			if idx, ok := trailingIndex(key); ok {
				names[idx] = val
			}
		case strings.HasPrefix(upper, "IPADDRESS"):
			if idx, ok := trailingIndex(key); ok {
				ips[idx] = val
			}
		case strings.HasPrefix(upper, "PORT"):
			if idx, ok := trailingIndex(key); ok {
				ports[idx] = val
			}
		}
	}

	var hosts []gotdx.HostInfo
	for idx := 1; idx <= 99; idx++ {
		ip, ok := ips[idx]
		if !ok || ip == "" {
			continue
		}
		port, err := strconv.Atoi(ports[idx])
		if err != nil || port <= 0 {
			continue
		}
		name := names[idx]
		if name == "" {
			name = fmt.Sprintf("%s %02d", section, idx)
		}
		hosts = append(hosts, gotdx.HostInfo{Name: name, IP: ip, Port: port})
	}
	if len(hosts) == 0 {
		return nil, fmt.Errorf("no hosts found in section %q", section)
	}
	return hosts, nil
}

// trailingIndex extracts the numeric suffix from keys like "IPAddress01".
func trailingIndex(key string) (int, bool) {
	i := len(key)
	for i > 0 && key[i-1] >= '0' && key[i-1] <= '9' {
		i--
	}
	if i == len(key) {
		return 0, false
	}
	idx, err := strconv.Atoi(key[i:])
	if err != nil {
		return 0, false
	}
	return idx, true
}
