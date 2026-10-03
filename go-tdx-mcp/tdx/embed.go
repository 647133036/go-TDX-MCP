package tdx

import _ "embed"

// embeddedConnectCfg is the official connect.cfg shipped with the TDX desktop
// client (extracted from the Linux .deb package). It carries the current
// production server list and is used so the collector probes the same hosts
// the desktop client does, without requiring the user to place a file on disk.
//
//go:embed connect.cfg
var embeddedConnectCfg []byte
