package tdx

import (
	"testing"

	"github.com/bensema/gotdx"
)

func TestParseConnectCfgHosts(t *testing.T) {
	hosts, err := ParseConnectCfgHosts("connect.cfg", "HQHOST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 38 {
		t.Fatalf("expected 38 HQ hosts from official client config, got %d", len(hosts))
	}
	first := hosts[0]
	if first.IP != "110.41.147.114" || first.Port != 7709 {
		t.Errorf("first host = %+v", first)
	}
	// GBK-decoded name should contain Chinese characters, not mojibake
	if first.Name == "" {
		t.Errorf("first host name is empty")
	}
	for _, h := range hosts {
		if h.IP == "" || h.Port <= 0 {
			t.Errorf("bad host entry: %+v", h)
		}
	}
}

func TestParseConnectCfgHostsMissingSection(t *testing.T) {
	if _, err := ParseConnectCfgHosts("connect.cfg", "NO_SUCH_SECTION"); err == nil {
		t.Fatalf("expected error for missing section")
	}
}

func TestParseConnectCfgHostsMissingFile(t *testing.T) {
	if _, err := ParseConnectCfgHosts("testdata/nonexistent.cfg", "HQHOST"); err == nil {
		t.Fatalf("expected error for missing file")
	}
}

func TestLoadConnectCfgHostsSections(t *testing.T) {
	hosts, err := LoadConnectCfgHosts("connect.cfg", DefaultSections)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 40 { // 38 from [HQHOST] + 2 from [HFHost]
		t.Fatalf("expected 40 hosts from official sections, got %d", len(hosts))
	}
	for _, h := range hosts {
		if h.Port != 7709 {
			t.Errorf("expected standard TDX port 7709, got %+v", h)
		}
	}
}

func TestLoadConnectCfgHostsDefaultsToSections(t *testing.T) {
	hosts, err := LoadConnectCfgHosts("connect.cfg", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 40 {
		t.Fatalf("expected default sections to yield 40 hosts, got %d", len(hosts))
	}
}

func TestLoadConnectCfgHostsBundledConfig(t *testing.T) {
	hosts, err := LoadConnectCfgHosts("", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 40 {
		t.Fatalf("expected bundled connect.cfg to yield 40 hosts, got %d", len(hosts))
	}
}

func TestEmbeddedConnectCfgHosts(t *testing.T) {
	hosts, err := EmbeddedConnectCfgHosts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) == 0 {
		t.Fatalf("expected at least one official host")
	}
}

func TestLoadConnectCfgHostsNoSectionsMatch(t *testing.T) {
	if _, err := LoadConnectCfgHosts("connect.cfg", []string{"NO_SUCH_SECTION"}); err == nil {
		t.Fatalf("expected error when no section matches")
	}
}

func TestMainHostsWithConnectCfg(t *testing.T) {
	merged := MainHostsWithConnectCfg()
	if len(merged) <= len(gotdx.MainHosts()) {
		t.Fatalf("expected merged list (%d) to grow beyond built-ins (%d)", len(merged), len(gotdx.MainHosts()))
	}
	seen := make(map[string]bool, len(merged))
	for _, h := range merged {
		if seen[h.Address()] {
			t.Errorf("duplicate host %s", h.Address())
			break
		}
		seen[h.Address()] = true
	}
	if merged[0].IP != "110.41.147.114" {
		t.Errorf("official config host should lead the merged list, got %+v", merged[0])
	}
}

func TestMergeHosts(t *testing.T) {
	base := []gotdx.HostInfo{
		{Name: "a", IP: "1.1.1.1", Port: 7709},
		{Name: "b", IP: "2.2.2.2", Port: 7709},
	}
	extra := []gotdx.HostInfo{
		{Name: "dup", IP: "1.1.1.1", Port: 7709},
		{Name: "c", IP: "3.3.3.3", Port: 7709},
	}
	merged := mergeHosts(base, extra)
	if len(merged) != 3 {
		t.Fatalf("expected 3 merged hosts, got %d", len(merged))
	}
	// extra takes precedence (placed first)
	if merged[0].Name != "dup" || merged[1].Name != "c" || merged[2].Name != "b" {
		t.Errorf("merged order wrong: %+v", merged)
	}
}
