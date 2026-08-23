package tdx

import (
	"testing"
)

func TestGetQuotesList(t *testing.T) {
	c := NewTDXTCPClient(10)
	reply, err := c.GetQuotesList(0, 0, 5)
	if err != nil {
		t.Fatalf("GetQuotesList error: %v", err)
	}
	t.Logf("Count=%d, List_len=%d", reply.Count, len(reply.List))
	for i, item := range reply.List {
		if i >= 5 {
			break
		}
		t.Logf("  [%d] code=%s, close=%.2f, pre_close=%.2f", i, item.Code, item.Close, item.PreClose)
	}
}