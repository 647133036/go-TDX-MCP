package tdx

import (
	"testing"
)

func TestSecurityList(t *testing.T) {
	c := NewTDXTCPClient(5)
	t.Run("SZ", func(t *testing.T) {
		reply, err := c.GetSecurityList(0, 0)
		if err != nil {
			t.Logf("SZ error: %v", err)
			return
		}
		t.Logf("SZ: count=%d, list_len=%d", reply.Count, len(reply.List))
		for i, sec := range reply.List {
			if i >= 5 {
				break
			}
			t.Logf("  [%d] code=%s, name=%s", i, sec.Code, sec.Name)
		}
	})
}