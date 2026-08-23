package tdx

import (
	"fmt"
	"testing"
)

func TestTCPNewEndpoints(t *testing.T) {
	c := NewTDXTCPClient(5)
	code := "000001"
	market := 0

	t.Run("GetVolumeProfile", func(t *testing.T) {
		reply, err := c.GetVolumeProfile(code, market)
		if err != nil {
			t.Logf("GetVolumeProfile error: %v (may be expected if not trading hours)", err)
			return
		}
		if reply == nil {
			t.Fatal("reply is nil")
		}
		t.Logf("VolumeProfile: code=%s, close=%.2f, profiles=%d", reply.Code, reply.Close, reply.Count)
	})

	t.Run("GetIndexInfo", func(t *testing.T) {
		reply, err := c.GetIndexInfo("000001", 1)
		if err != nil {
			t.Logf("GetIndexInfo error: %v", err)
			return
		}
		if reply == nil {
			t.Fatal("reply is nil")
		}
		t.Logf("IndexInfo: code=%s, close=%.2f, orders=%d, up=%d, down=%d", reply.Code, reply.Close, reply.OrderCount, reply.UpCount, reply.DownCount)
	})

	t.Run("GetIndexMomentum", func(t *testing.T) {
		reply, err := c.GetIndexMomentum("000001", 1)
		if err != nil {
			t.Logf("GetIndexMomentum error: %v", err)
			return
		}
		if reply == nil {
			t.Fatal("reply is nil")
		}
		t.Logf("IndexMomentum: count=%d, values_count=%d", reply.Count, len(reply.Values))
	})

	t.Run("GetHistoryOrders", func(t *testing.T) {
		_, err := c.GetHistoryOrders(20260821, "000001", 0)
		if err != nil {
			t.Logf("GetHistoryOrders error: %v (may be expected for non-trading day)", err)
			return
		}
		t.Log("GetHistoryOrders: success")
	})

	t.Run("GetXDXRInfo", func(t *testing.T) {
		reply, err := c.GetXDXRInfo(code, market)
		if err != nil {
			t.Logf("GetXDXRInfo error: %v", err)
			return
		}
		if reply == nil {
			t.Fatal("reply is nil")
		}
		t.Logf("XDXR: code=%s, count=%d", reply.Code, reply.Count)
	})

	t.Run("GetMACCapitalFlow", func(t *testing.T) {
		reply, err := c.GetMACCapitalFlow(code, market)
		if err != nil {
			t.Logf("GetMACCapitalFlow error: %v", err)
			return
		}
		if reply == nil {
			t.Fatal("reply is nil")
		}
		t.Logf("MACCapitalFlow: main_in=%.0f, main_out=%.0f", reply.TodayMainIn, reply.TodayMainOut)
	})

	t.Run("ExGetQuotesList", func(t *testing.T) {
		reply, err := c.ExGetQuotesList(31, 0, false, 0, 5)
		if err != nil {
			t.Logf("ExGetQuotesList error: %v", err)
			return
		}
		if reply == nil {
			t.Fatal("reply is nil")
		}
		t.Logf("ExGetQuotesList: count=%d, list_len=%d", reply.Count, len(reply.List))
		if len(reply.List) > 0 {
			item := reply.List[0]
			fmt.Printf("  first item: code=%s, close=%.2f\n", item.Code, item.Close)
		}
	})
}